// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"iter"

	"go.wdy.de/nago/auth"
)

// Aggregate is a mutable, event-sourced aggregate root.
//
// It must be a pointer type. Evolve (on the events) mutates it in place; the
// handler keeps one live instance per aggregate id in memory and hands out deep
// copies via Clone for reads.
type Aggregate[T any] interface {
	// Clone returns a deep copy that shares no mutable memory with the receiver.
	Clone() T

	// IsDeleted reports whether the aggregate has been semantically deleted. The
	// handler checks this after applying each event: once true, the aggregate is
	// dropped from the in-memory set (and stays gone on replay), so a deleted
	// aggregate never lingers and its id may even be reused by a later create.
	IsDeleted() bool
}

// AggregateID extracts the aggregate key an event belongs to. The boolean is
// false when the event belongs to no aggregate (e.g. a cross-cutting fact); such
// events are skipped by the handler's routing. It replaces the old persistent
// primary index: the handler rebuilds all aggregates by replaying every event
// once and routing each via this function.
type AggregateID[E any, Primary ~string] func(E) (Primary, bool)

// Backend is the storage seam the [Handler] sits on. It is deliberately narrow
// so the same handler runs on different engines (a blob-backed store, the ndb
// message store, ...). An implementation owns serialization, sequence
// assignment and the physical layout; the handler stays oblivious to all of it.
type Backend[E any] interface {
	// Append durably stores e and returns the populated envelope (with the
	// assigned global Sequence and the store's append time).
	Append(subject auth.Subject, e E) (Envelope[E], error)

	// ReplayAll yields every stored event in ascending global Sequence order.
	// This is the single read path the eager handler uses to rebuild state on
	// startup; there is no per-aggregate index.
	ReplayAll(subject auth.Subject) iter.Seq2[Envelope[E], error]
}

// Cmd is a command which also provides the Decide implementation.
type Cmd[Aggregate, SuperEvt any] interface {
	Decide(auth.Subject, Aggregate) ([]SuperEvt, error)
}

// Evt is an event which also provides the Evolve implementation and a stable
// discriminator for (de)serialization.
type Evt[Aggregate any] interface {
	Evolve(context.Context, Aggregate) error
	Discriminator() Discriminator
}
