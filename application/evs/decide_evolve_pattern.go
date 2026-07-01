// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"reflect"
	"sync"
	"sync/atomic"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

// aggregateState holds one aggregate in two copies to make reads cheap and
// race-free (an RCU / read-copy-update style):
//
//   - live is the internal source of truth. Only the handler mutates it, always
//     under mutex, by folding events with Evolve. Readers never touch it.
//   - snapshot is a lazily produced deep clone handed to readers lock-free via
//     an atomic pointer. It is (re)built on the first read after a change and
//     reused by further reads until the next write marks it stale.
//
// Because readers only ever see the snapshot, an accidental mutation on a
// returned aggregate can never corrupt the truth: the next write evolves live
// and the next read replaces the snapshot with a fresh clone.
type aggregateState[Aggregate any] struct {
	mutex    sync.Mutex
	live     Aggregate
	snapshot atomic.Pointer[Aggregate]
	stale    atomic.Bool
}

// Handler is the default implementation of the decide-evolve pattern.
//
// It is eager: on first use it replays the entire event log once (via
// [Backend.ReplayAll]), routes each event to its aggregate using the injected
// [AggregateID] extractor, and folds it in with Evolve. All aggregates are then
// kept in memory; there is no per-aggregate index and no lazy per-aggregate
// replay. This trades memory (every aggregate stays resident) for simplicity and
// speed, and suits domains with a bounded number of aggregates.
//
// An aggregate that reports IsDeleted after an event is dropped from the set, so
// deletions are ordinary events that vanish correctly on replay.
type Handler[Aggregate aggregateConstraint[Aggregate], SuperEvt Evt[Aggregate], Primary ~string] struct {
	backend    Backend[SuperEvt]
	aggID      AggregateID[SuperEvt, Primary]
	eventTypes map[Discriminator]func() SuperEvt
	register   Register[SuperEvt]

	aggregates concurrent.RWMap[Primary, *aggregateState[Aggregate]]
	initOnce   sync.Once
	initErr    error
}

// aggregateConstraint binds the aggregate type to the [Aggregate] contract.
type aggregateConstraint[T any] interface {
	Aggregate[T]
}

// NewHandler creates a handler backed by the given [Backend], using aggID to
// route events to their aggregate. register associates event Go types with their
// discriminators for the backend's (de)serialization; pass uc.Register from a
// wired backend, or nil if the backend does not need a registry.
func NewHandler[Aggregate aggregateConstraint[Aggregate], SuperEvt Evt[Aggregate], Primary ~string](
	backend Backend[SuperEvt],
	aggID AggregateID[SuperEvt, Primary],
	register Register[SuperEvt],
) *Handler[Aggregate, SuperEvt, Primary] {
	if reflect.TypeFor[Aggregate]().Kind() != reflect.Ptr {
		panic("Aggregate must be a pointer type")
	}

	return &Handler[Aggregate, SuperEvt, Primary]{
		backend:    backend,
		aggID:      aggID,
		register:   register,
		eventTypes: make(map[Discriminator]func() SuperEvt),
	}
}

// RegisterEvents declares the event types this handler manages. It is not thread
// safe and must be called before any other operation. It validates the
// discriminator, registers the type (if a registry is present), checks JSON
// marshalability, and builds the local discriminator→factory map.
func (h *Handler[Aggregate, SuperEvt, Primary]) RegisterEvents(m ...Evt[Aggregate]) {
	for _, e := range m {
		if err := e.Discriminator().Validate(); err != nil {
			panic(fmt.Errorf("type %T has invalid discriminator: %w", e, err))
		}

		if t, ok := h.eventTypes[e.Discriminator()]; ok {
			panic(fmt.Errorf("type %T and %T share the same discriminator %s but must be unique", t, e, e.Discriminator()))
		}

		rType := reflect.TypeOf(e)
		if rType.Kind() == reflect.Ptr {
			rType = rType.Elem()
		}

		if h.register != nil {
			if err := h.register(rType, e.Discriminator()); err != nil {
				panic(fmt.Errorf("cannot register type %T: %w", e, err))
			}
		}

		h.eventTypes[e.Discriminator()] = func() SuperEvt {
			return reflect.New(rType).Elem().Interface().(SuperEvt)
		}
	}
}

// ensureInit replays the whole log exactly once and builds the in-memory
// aggregate set. Safe for concurrent callers.
func (h *Handler[Aggregate, SuperEvt, Primary]) ensureInit() error {
	h.initOnce.Do(func() {
		ctx := context.Background()
		for env, err := range h.backend.ReplayAll(user.SU()) {
			if err != nil {
				h.initErr = fmt.Errorf("cannot replay events: %w", err)
				return
			}

			key, ok := h.aggID(env.Data)
			if !ok {
				continue // event belongs to no aggregate
			}

			st := h.loadOrCreate(key)
			if err := env.Data.Evolve(ctx, st.live); err != nil {
				h.initErr = fmt.Errorf("cannot evolve aggregate %v: %w: type %T", key, err, env.Data)
				return
			}

			if st.live.IsDeleted() {
				h.aggregates.Delete(key)
			}
		}
	})

	return h.initErr
}

func (h *Handler[Aggregate, SuperEvt, Primary]) loadOrCreate(key Primary) *aggregateState[Aggregate] {
	st, loaded := h.aggregates.LoadOrStore(key, &aggregateState[Aggregate]{
		live: reflect.New(reflect.TypeFor[Aggregate]().Elem()).Interface().(Aggregate),
	})
	if !loaded {
		// fresh state: no snapshot yet, so the first read must clone.
		st.stale.Store(true)
	}
	return st
}

// Handle runs a command against the current aggregate state: decide → append →
// evolve. Generated events are stored first (improving the chance the durable
// log and in-memory state agree), then applied as a batch. If the aggregate
// reports IsDeleted after an event, it is dropped from the set.
func (h *Handler[Aggregate, SuperEvt, Primary]) Handle(subject auth.Subject, key Primary, cmd Cmd[Aggregate, SuperEvt]) error {
	if key == "" {
		return fmt.Errorf("aggregate id / key cannot be empty")
	}

	if err := h.ensureInit(); err != nil {
		return err
	}

	st := h.loadOrCreate(key)
	st.mutex.Lock()
	defer st.mutex.Unlock()

	events, err := cmd.Decide(subject, st.live)
	if err != nil {
		return fmt.Errorf("cannot decide command %T: %w", cmd, err)
	}

	if len(events) == 0 {
		return nil
	}

	batch := make([]SuperEvt, 0, len(events))
	for _, e := range events {
		if _, err := h.backend.Append(user.SU(), e); err != nil {
			return fmt.Errorf("cannot store event %T: %w", e, err)
		}
		batch = append(batch, e)
	}

	for _, e := range batch {
		if e2 := e.Evolve(subject.Context(), st.live); e2 != nil {
			// in-memory state is now partially applied and inconsistent; drop it
			// (and its snapshot) so the next access rebuilds from the durable log.
			st.snapshot.Store(nil)
			h.aggregates.Delete(key)
			return fmt.Errorf("cannot evolve aggregate: %w: type %T", e2, e)
		}
	}

	if st.live.IsDeleted() {
		st.snapshot.Store(nil)
		h.aggregates.Delete(key)
		return nil
	}

	// the live state changed: the reader snapshot is now out of date and will be
	// re-cloned lazily on the next read.
	st.stale.Store(true)
	return nil
}

// All returns all live aggregate ids. Because every aggregate is resident, this
// is a pure in-memory enumeration of the aggregate set.
func (h *Handler[Aggregate, SuperEvt, Primary]) All(ctx context.Context) (iter.Seq[Primary], error) {
	if err := h.ensureInit(); err != nil {
		return nil, err
	}

	return func(yield func(Primary) bool) {
		for key := range h.aggregates.All() {
			if !yield(key) {
				return
			}
		}
	}, nil
}

// Aggregate returns the current aggregate snapshot, or [os.ErrNotExist] if no
// such (live) aggregate exists.
//
// The returned value is a deep clone that is safe to read concurrently with
// writes: it is served lock-free from an atomic snapshot pointer and is only
// (re)cloned on the first read after a change. Repeated reads without an
// intervening write return the same snapshot without cloning again.
//
// Callers must treat the result as read-only. Mutating it is harmless to the
// store (the truth is a separate instance) but pointless — the change is not
// persisted and is discarded on the next re-clone.
func (h *Handler[Aggregate, SuperEvt, Primary]) Aggregate(ctx context.Context, key Primary) (Aggregate, error) {
	if err := h.ensureInit(); err != nil {
		var zero Aggregate
		return zero, err
	}

	st, ok := h.aggregates.Get(key)
	if !ok {
		var zero Aggregate
		return zero, fmt.Errorf("aggregate %v: %w", key, os.ErrNotExist)
	}

	// fast path: a fresh snapshot exists — no lock, no clone.
	if !st.stale.Load() {
		if snap := st.snapshot.Load(); snap != nil {
			return *snap, nil
		}
	}

	// slow path: (re)clone the live state under the lock.
	st.mutex.Lock()
	defer st.mutex.Unlock()

	// re-check under the lock: another goroutine may have refreshed it.
	if st.stale.Load() || st.snapshot.Load() == nil {
		clone := st.live.Clone()
		st.snapshot.Store(&clone)
		st.stale.Store(false)
	}
	return *st.snapshot.Load(), nil
}

// Replay yields the raw events of a single aggregate by filtering the global
// replay through the aggregate id extractor. It does not mutate any state and is
// intended for debugging, export or backup.
func (h *Handler[Aggregate, SuperEvt, Primary]) Replay(key Primary) iter.Seq2[Envelope[SuperEvt], error] {
	return func(yield func(Envelope[SuperEvt], error) bool) {
		for env, err := range h.backend.ReplayAll(user.SU()) {
			if err != nil {
				if !yield(env, err) {
					return
				}
				continue
			}

			k, ok := h.aggID(env.Data)
			if !ok || k != key {
				continue
			}

			if !yield(env, nil) {
				return
			}
		}
	}
}

// Delete semantically deletes an aggregate by issuing cmd (which is expected to
// produce a deletion event whose Evolve sets IsDeleted). It is a thin wrapper
// over [Handler.Handle] kept for call-site clarity.
func (h *Handler[Aggregate, SuperEvt, Primary]) Delete(subject auth.Subject, key Primary, cmd Cmd[Aggregate, SuperEvt]) error {
	return h.Handle(subject, key, cmd)
}

// Restore blindly decodes and re-stores the given raw envelopes. It is dangerous
// (it can corrupt aggregate state) and is meant only for restoring or migrating
// event data. After a restore the in-memory set is reset so it rebuilds lazily.
func (h *Handler[Aggregate, SuperEvt, Primary]) Restore(ctx context.Context, it iter.Seq2[JsonEnvelope, error]) error {
	registry := make(map[Discriminator]reflect.Type)
	for d, factory := range h.eventTypes {
		registry[d] = reflect.TypeOf(factory())
	}

	var decoded []SuperEvt
	for env, err := range it {
		if err != nil {
			return err
		}

		obj, err := env.Decode(registry)
		if err != nil {
			return err
		}

		e, ok := obj.(SuperEvt)
		if !ok {
			return fmt.Errorf("decoded event %T is not a SuperEvt", obj)
		}
		decoded = append(decoded, e)
	}

	if len(decoded) == 0 {
		return nil
	}

	for _, e := range decoded {
		if _, err := h.backend.Append(user.SU(), e); err != nil {
			return err
		}
	}

	slog.Info("restored events", "count", len(decoded))
	h.reset()
	return nil
}

// reset drops the in-memory set and arms a fresh replay on next access.
func (h *Handler[Aggregate, SuperEvt, Primary]) reset() {
	h.aggregates.Clear()
	h.initOnce = sync.Once{}
	h.initErr = nil
}
