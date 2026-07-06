// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"encoding/json"
	"fmt"
	"iter"
	"math"
	"reflect"
	"time"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xtime"
)

// ndbBackend implements [Backend] on top of an [ndb.Messages] engine.
//
// Each event type's discriminator is used directly as the [ndb.TypeID] (both are
// stable textual identifiers and share the same allowed character set), so the
// store fans events of one type into their own stream. The event's own JSON is
// stored verbatim as the opaque message payload — no wrapper envelope — keeping
// the engine schema-free. The envelope's read-side EventTime is derived from the
// engine's native append time (msg.TimeNano).
type ndbBackend[E Evt[A], A any] struct {
	msgs ndb.Messages

	byType        concurrent.RWMap[reflect.Type, Discriminator]
	byDiscrimentr concurrent.RWMap[Discriminator, reflect.Type]
}

var _ RegisterableBackend[Evt[any]] = (*ndbBackend[Evt[any], any])(nil)

// NewNDBBackend returns a [Backend] persisting events into an [ndb.Messages]
// engine. It performs no permission checks of its own; the handler drives it as
// the super user.
func NewNDBBackend[E Evt[A], A any](msgs ndb.Messages) RegisterableBackend[E] {
	return &ndbBackend[E, A]{msgs: msgs}
}

// Register associates a Go type with a discriminator for (de)serialization.
func (b *ndbBackend[E, A]) Register(t reflect.Type, d Discriminator) error {
	if err := d.Validate(); err != nil {
		return err
	}
	if !ndb.ValidTypeID(string(d)) {
		return fmt.Errorf("discriminator %q is not a valid ndb type id", d)
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

func (b *ndbBackend[E, A]) Append(subject auth.Subject, e E) (Envelope[E], error) {
	var zero Envelope[E]

	discriminator, ok := b.byType.Get(reflect.TypeOf(e))
	if !ok {
		return zero, fmt.Errorf("type %T not registered; call RegisterEvents first", e)
	}

	dataBuf, err := json.Marshal(e)
	if err != nil {
		return zero, fmt.Errorf("event %T cannot be marshalled: %w", e, err)
	}

	seq, err := b.msgs.Append(ndb.TypeID(discriminator), ndb.NewTraceID(), dataBuf)
	if err != nil {
		return zero, fmt.Errorf("error appending event: %w", err)
	}

	key, _ := NewSeqKey(SeqID(seq))
	return Envelope[E]{
		Sequence:      SeqID(seq),
		Key:           key,
		Discriminator: discriminator,
		EventTime:     xtime.Now(),
		Data:          e,
		Raw:           dataBuf,
	}, nil
}

// ReplayAll yields this backend's events in ascending global Sequence order. It
// reads ONLY the registered discriminators, never the whole store: an
// [ndb.Messages] engine may be shared by several aggregates (a single engine per
// bounded context or per process), each backend writing its own event types into
// the same stream. Passing the registered type set to [ndb.History.Replay] keeps
// each backend from decoding a sibling aggregate's events (which would fail with
// an "unknown discriminator" and abort the handler's replay). This mirrors how a
// [Projection] hands its type set to [ndb.Tail].
func (b *ndbBackend[E, A]) ReplayAll(subject auth.Subject) iter.Seq2[Envelope[E], error] {
	// Snapshot the registered discriminators as the type filter.
	types := make([]ndb.TypeID, 0, b.byDiscrimentr.Len())
	for d := range b.byDiscrimentr.All() {
		types = append(types, ndb.TypeID(d))
	}

	return func(yield func(Envelope[E], error) bool) {
		// No registered types means there is nothing this backend could decode.
		// Return empty instead of calling Replay with an empty slice, which the
		// contract interprets as "all types" — exactly the over-read this method
		// must avoid.
		if len(types) == 0 {
			return
		}

		var zero Envelope[E]
		for typeID, msg := range b.msgs.Replay(types, 1, ndb.Seq(math.MaxUint64)) {
			env, err := b.decode(Discriminator(typeID), msg)
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

func (b *ndbBackend[E, A]) decode(discriminator Discriminator, msg ndb.Message) (Envelope[E], error) {
	var zero Envelope[E]

	if msg.Encoding != ndb.EncodingRaw {
		return zero, fmt.Errorf("unexpected payload encoding %d for %q", msg.Encoding, discriminator)
	}

	rtype, ok := b.byDiscrimentr.Get(discriminator)
	if !ok {
		return zero, fmt.Errorf("unknown discriminator %q", discriminator)
	}

	rval := reflect.New(rtype)
	if err := json.Unmarshal(msg.Payload, rval.Interface()); err != nil {
		return zero, err
	}

	evt, ok := rval.Elem().Interface().(E)
	if !ok {
		return zero, fmt.Errorf("type mismatch for discriminator %q: stored type is not convertible into %T (incompatible refactor?)", discriminator, *new(E))
	}

	// Payload is a view into the engine's reusable read buffer; clone it so the
	// envelope can be retained beyond this iteration step.
	raw := make([]byte, len(msg.Payload))
	copy(raw, msg.Payload)

	key, _ := NewSeqKey(SeqID(msg.Seq))
	return Envelope[E]{
		Sequence:      SeqID(msg.Seq),
		Key:           key,
		Discriminator: discriminator,
		EventTime:     xtime.UnixMilliseconds(msg.TimeNano / int64(time.Millisecond)),
		Data:          evt,
		Raw:           raw,
	}, nil
}
