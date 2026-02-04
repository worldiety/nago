// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"reflect"
	"sync"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

// Cmd is an interface for a command which also provides the Decide implementation.
type Cmd[Aggregate, SuperEvt any] interface {
	Decide(auth.Subject, Aggregate) ([]SuperEvt, error)
}

// Evt is an interface for an event which also provides the Evolve implementation.
type Evt[Aggregate any] interface {
	Evolve(context.Context, Aggregate) error
	Discriminator() Discriminator
}

// Cloner declares a contract which performs a deep copy of a mutable aggregate.
type Cloner[T any] interface {
	// Clone returns a deep copy of itself. An implementation must not return mutable memory shared with
	// any clone.
	Clone() T
}

type aggregateState[Aggregate any] struct {
	replayed  bool
	aggregate Aggregate
	mutex     sync.Mutex
}

// Handler provides a default implementation for the decide-evolve pattern by using the basic use-cases of
// this package. It trades the simplicity of a mutable and immutable aggregate for performance and thread safety
// against a per-aggregate mutex and keeping the aggregate in memory and creates clones under the same lock
// for reading.
type Handler[Aggregate Cloner[Aggregate], SuperEvt Evt[Aggregate], Primary ~string] struct {
	uc             UseCases[SuperEvt]
	eventTypes     map[Discriminator]func() SuperEvt
	replay         ReplayWithIndex[Primary, SuperEvt]
	storeEvent     Store[SuperEvt]
	register       Register[SuperEvt]
	aggregateCache concurrent.RWMap[Primary, *aggregateState[Aggregate]]
	index          *StoreIndex[Primary, SuperEvt]
}

func NewHandler[Aggregate Cloner[Aggregate], SuperEvt Evt[Aggregate], Primary ~string](
	uc UseCases[SuperEvt],
	replay ReplayWithIndex[Primary, SuperEvt],
	index *StoreIndex[Primary, SuperEvt],
) *Handler[Aggregate, SuperEvt, Primary] {
	h := &Handler[Aggregate, SuperEvt, Primary]{
		replay:     replay,
		uc:         uc,
		storeEvent: uc.Store,
		register:   uc.Register,
		eventTypes: make(map[Discriminator]func() SuperEvt),
		index:      index,
	}

	if reflect.TypeFor[Aggregate]().Kind() != reflect.Ptr {
		panic("Aggregate must be a pointer type")
	}

	return h
}

// RegisterEvents is not thread safe and must be called before any other operation.
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

		if err := h.register(rType, e.Discriminator()); err != nil {
			panic(fmt.Errorf("cannot register type %T: %w", e, err))
		}

		// check if marshaling works
		if _, err := json.Marshal(e); err != nil {
			panic(fmt.Errorf("cannot marshal type %T: %w", e, err))
		}

		h.eventTypes[e.Discriminator()] = func() SuperEvt {
			return reflect.New(rType).Elem().Interface().(SuperEvt)
		}

	}

}

// ensureReplayed expects to be run on active mutex lock.
func (h *Handler[Aggregate, SuperEvt, Primary]) ensureReplayed(ctx context.Context, state *aggregateState[Aggregate], key Primary) error {

	if !state.replayed {
		h.resetAggregate(state)
		count := 0
		for env, err := range h.replay(user.SU(), key, ReplayOptions{}) {
			if err != nil {
				return fmt.Errorf("cannot replay events: %w", err)
			}

			// implementation note: this works, because we ensured in Handler constructor
			// that mutAggregate is a pointer type
			if err := env.Data.Evolve(ctx, state.aggregate); err != nil {
				return fmt.Errorf("cannot evolve aggregate: %w: type %T", err, env.Data)
			}
			count++
		}

		if count == 0 {
			return fmt.Errorf("no events found for aggregate %s: %w", key, os.ErrNotExist)
		}

		state.replayed = true
	}

	return nil
}

// resetAggregate expects to be run on active mutex lock.
func (h *Handler[Aggregate, SuperEvt, Primary]) resetAggregate(state *aggregateState[Aggregate]) {
	state.replayed = false
	state.aggregate = reflect.New(reflect.TypeFor[Aggregate]().Elem()).Interface().(Aggregate)
}

func (h *Handler[Aggregate, SuperEvt, Primary]) Delete(subject auth.Subject, key Primary) error {
	if err := h.uc.DeleteByPrimary(subject, h.index.Info().ID, string(key)); err != nil {
		return err
	}

	h.Evict(key)
	return nil
}

func (h *Handler[Aggregate, SuperEvt, Primary]) Handle(subject auth.Subject, key Primary, cmd Cmd[Aggregate, SuperEvt]) error {
	if key == "" {
		return fmt.Errorf("aggregate id / key cannot be empty")
	}

	state, _ := h.aggregateCache.LoadOrStore(key, &aggregateState[Aggregate]{})
	state.mutex.Lock()
	defer state.mutex.Unlock()

	if err := h.ensureReplayed(subject.Context(), state, key); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		// ignore not-exist errors, may be a bootstrap command
	}

	events, err := cmd.Decide(subject, state.aggregate)
	if err != nil {
		return fmt.Errorf("cannot decide command %T: %w", cmd, err)
	}

	// fast return: Decide did not produce any events. This optimization must always be "free"
	if len(events) == 0 {
		return nil
	}

	batch := make([]Envelope[SuperEvt], 0, len(events))
	// first store all generated events to improve consistency probability slightly
	for _, e := range events {
		env, err := h.storeEvent(user.SU(), e, StoreOptions{
			CreatedBy: subject.ID(),
		})

		if err != nil {
			// this is less likely to happen and only occurs if storage is full.
			// We already checked if Evt can get marshaled at register time.
			return fmt.Errorf("cannot store event %T: %w", e, err)
		}

		batch = append(batch, env)
	}

	// second: apply them as an atomic batch
	for _, env := range batch {
		if e2 := env.Data.Evolve(subject.Context(), state.aggregate); e2 != nil {
			err = fmt.Errorf("cannot evolve aggregate: %w: type %T", e2, env.Data)
			break
		}
	}

	if err != nil {
		// our aggregate is now corrupted and partially applied. Reset the mutable aggregate for the next read to
		// be consistent again
		h.resetAggregate(state)
		return err
	}

	return nil
}

// All returns all aggregate ids. The implementation avoids to load replay all aggregates and just consults
// the inverse primary index to return potential aggregates.
func (h *Handler[Aggregate, SuperEvt, Primary]) All(ctx context.Context) (iter.Seq[Primary], error) {
	it, err := h.index.GroupPrimary(ctx)
	if err != nil {
		return nil, err
	}

	return func(yield func(Primary) bool) {
		for id := range it {
			if !yield(id) {
				return
			}
		}
	}, nil
}

// Aggregate returns the current immutable aggregate snapshot. This is only race-free if ReadOnlyAggregate does
// not share mutable memory with the mutable aggregate.
func (h *Handler[Aggregate, SuperEvt, Primary]) Aggregate(ctx context.Context, key Primary) (Aggregate, error) {
	state, _ := h.aggregateCache.LoadOrStore(key, &aggregateState[Aggregate]{})
	state.mutex.Lock()
	defer state.mutex.Unlock()

	if err := h.ensureReplayed(ctx, state, key); err != nil {
		var zero Aggregate
		return zero, err
	}

	return state.aggregate.Clone(), nil
}

// Evict removes the aggregate from the cache. This must be done manually.
func (h *Handler[Aggregate, SuperEvt, Primary]) Evict(key Primary) {
	h.aggregateCache.Delete(key)
}

func (h *Handler[Aggregate, SuperEvt, Primary]) EvictAll() {
	h.aggregateCache.Clear()
}

// Replay replays all events for the given aggregate id in raw form. It does not mutate any internal state.
// This is useful for debugging or backup purposes.
func (h *Handler[Aggregate, SuperEvt, Primary]) Replay(key Primary) iter.Seq2[Envelope[SuperEvt], error] {
	return h.replay(user.SU(), key, ReplayOptions{})
}

// Restore is in the danger zone of the handler because it blindly tries to decode the given envelope and
// evolves the affected aggregate. It may screw up the internal state of any aggregate. Its use is mainly for
// restoring or exchanging aggregate states.
func (h *Handler[Aggregate, SuperEvt, Primary]) Restore(ctx context.Context, it iter.Seq2[JsonEnvelope, error]) error {
	registry := make(map[Discriminator]reflect.Type)
	for registeredType := range h.uc.RegisteredTypes() {
		registry[registeredType.Discriminator] = registeredType.Type
	}

	var decodedEvents []Envelope[SuperEvt]
	for env, err := range it {
		if err != nil {
			return err
		}

		obj, err := env.Decode(registry)
		if err != nil {
			return err
		}

		if e, ok := obj.(SuperEvt); ok {
			decodedEvents = append(decodedEvents, Envelope[SuperEvt]{
				Discriminator: env.Discriminator,
				EventTime:     env.EventTime,
				CreatedBy:     env.CreatedBy,
				Metadata:      env.Metadata,
				Data:          e,
				Raw:           env.Data,
			})
		} else {
			return fmt.Errorf("decoded event %T is not a SuperEvt", obj)
		}
	}

	if len(decodedEvents) == 0 {
		return nil
	}

	for _, event := range decodedEvents {
		_, err := h.uc.Store(user.SU(), event.Data, StoreOptions{
			Metadata:  event.Metadata,
			EventTime: event.EventTime,
			CreatedBy: event.CreatedBy,
		})
		if err != nil {
			return err
		}
	}

	slog.Info("restored events", "count", len(decodedEvents))
	h.EvictAll()
	return nil
}
