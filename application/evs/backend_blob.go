// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"reflect"
	"sync"
	"sync/atomic"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xtime"
)

// blobBackend implements [Backend] on top of a single blob.Store. Events are
// stored as JSON envelopes keyed by a zero-padded sequence key; the next
// sequence is bootstrapped once by reading the highest existing key in reverse.
// It keeps its own discriminator↔type registry (populated via Register).
type blobBackend[E Evt[A], A any] struct {
	store blob.Store

	byType        concurrent.RWMap[reflect.Type, Discriminator]
	byDiscrimentr concurrent.RWMap[Discriminator, reflect.Type]

	lastID   atomic.Int64
	bootOnce sync.Once
	bootErr  error
}

// RegisterableBackend is a [Backend] that additionally exposes a Register hook
// matching the [Register] signature. It lets callers (e.g. cfgevs.NewHandler)
// wire event types into the backend's discriminator registry without knowing the
// concrete (unexported) backend type.
type RegisterableBackend[E any] interface {
	Backend[E]
	Register(t reflect.Type, d Discriminator) error
}

// NewBlobBackend returns a [Backend] persisting events into eventStore as JSON
// envelopes. It performs no permission checks of its own; the handler drives it
// as the super user.
func NewBlobBackend[E Evt[A], A any](eventStore blob.Store) RegisterableBackend[E] {
	return &blobBackend[E, A]{store: eventStore}
}

// Register associates a Go type with a discriminator for (de)serialization.
// It matches the [Register] signature so it can be passed to [NewHandler].
func (b *blobBackend[E, A]) Register(t reflect.Type, d Discriminator) error {
	if err := d.Validate(); err != nil {
		return err
	}
	if existing, ok := b.byType.Get(t); ok {
		if existing != d {
			return fmt.Errorf("type %v already registered as %q, not %q", t, existing, d)
		}
		return nil
	}
	if existing, ok := b.byDiscrimentr.Get(d); ok && existing != t {
		return fmt.Errorf("discriminator %q already registered for %v", d, existing)
	}
	b.byType.Put(t, d)
	b.byDiscrimentr.Put(d, t)
	return nil
}

func (b *blobBackend[E, A]) bootstrap() error {
	b.bootOnce.Do(func() {
		for sid, err := range b.store.List(context.Background(), blob.ListOptions{Reverse: true}) {
			if err != nil {
				b.bootErr = fmt.Errorf("error listing events: %w", err)
				return
			}
			n, err := SeqKey(sid).Parse()
			if err != nil {
				b.bootErr = fmt.Errorf("error parsing event key %s: %w", sid, err)
				return
			}
			b.lastID.Store(int64(n))
			break // only the highest key
		}
	})
	return b.bootErr
}

func (b *blobBackend[E, A]) Append(subject auth.Subject, e E) (Envelope[E], error) {
	var zero Envelope[E]

	if err := b.bootstrap(); err != nil {
		return zero, err
	}

	discriminator, ok := b.byType.Get(reflect.TypeOf(e))
	if !ok {
		return zero, fmt.Errorf("type %T not registered; call RegisterEvents first", e)
	}

	payloadBuf, err := json.Marshal(e)
	if err != nil {
		return zero, fmt.Errorf("event %T cannot be marshalled: %w", e, err)
	}

	seqID := SeqID(b.lastID.Add(1))
	key, err := NewSeqKey(seqID)
	if err != nil {
		return zero, err
	}

	// blob has no native append time, so we record it ourselves.
	eventTime := xtime.Now()

	env := JsonEnvelope{
		Discriminator: discriminator,
		EventTime:     eventTime,
		Data:          payloadBuf,
	}

	buf, err := json.Marshal(env)
	if err != nil {
		return zero, fmt.Errorf("error marshalling envelope: %w", err)
	}

	if err := blob.Put(b.store, string(key), buf); err != nil {
		return zero, fmt.Errorf("error storing envelope: %w", err)
	}

	return Envelope[E]{
		Sequence:      seqID,
		Key:           key,
		Discriminator: discriminator,
		EventTime:     eventTime,
		Data:          e,
		Raw:           payloadBuf,
	}, nil
}

func (b *blobBackend[E, A]) ReplayAll(subject auth.Subject) iter.Seq2[Envelope[E], error] {
	return func(yield func(Envelope[E], error) bool) {
		var zero Envelope[E]
		for key, err := range b.store.List(context.Background(), blob.ListOptions{}) {
			if err != nil {
				if !yield(zero, err) {
					return
				}
				continue
			}

			seqID, err := SeqKey(key).Parse()
			if err != nil {
				if !yield(zero, err) {
					return
				}
				continue
			}

			env, err := b.load(seqID, SeqKey(key))
			if err != nil {
				if !yield(zero, err) {
					return
				}
				continue
			}

			if !yield(env, nil) {
				return
			}
		}
	}
}

func (b *blobBackend[E, A]) load(id SeqID, key SeqKey) (Envelope[E], error) {
	var zero Envelope[E]

	optBuf, err := blob.Get(b.store, string(key))
	if err != nil {
		return zero, err
	}
	if optBuf.IsNone() {
		return zero, fmt.Errorf("event %s missing", key)
	}

	var jsonEnv JsonEnvelope
	if err := json.Unmarshal(optBuf.Unwrap(), &jsonEnv); err != nil {
		return zero, err
	}

	payload, err := jsonEnv.decodeData(&b.byDiscrimentr)
	if err != nil {
		return zero, err
	}

	evt, ok := payload.(E)
	if !ok {
		return zero, fmt.Errorf("type mismatch for discriminator %q: stored type is not convertible into %T (incompatible refactor?)", jsonEnv.Discriminator, *new(E))
	}

	return Envelope[E]{
		Sequence:      id,
		Key:           key,
		Discriminator: jsonEnv.Discriminator,
		EventTime:     jsonEnv.EventTime,
		Data:          evt,
		Raw:           jsonEnv.Data,
	}, nil
}
