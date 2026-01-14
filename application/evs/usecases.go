// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"fmt"
	"iter"
	"reflect"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xtime"
)

type Discriminator string

func (d Discriminator) Validate() error {
	if len(d) == 0 {
		return fmt.Errorf("discriminator must not be empty")
	}

	if len(d) > 256 {
		return fmt.Errorf("discriminator exceeds max length of 256: %d", len(d))
	}

	for i, r := range d {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.') {
			return fmt.Errorf("discriminator contains invalid character at position %d: %q (allowed: a-z, A-Z, 0-9, -, _, .)", i, r)
		}
	}

	return nil
}

// SeqID is a strict monotonic increasing sequence identifier, which is guaranteed to be unique per topic.
// It is similar to a Kafka offset, but technically it is more like a database autoincrement.
type SeqID int64

// Envelope uses explicitly unix milliseconds to avoid any questions and problems regarding with different timezones.
type Envelope[Evt any] struct {
	Key           SeqKey
	Sequence      SeqID
	Discriminator Discriminator
	EventTime     xtime.UnixMilliseconds
	CreatedBy     user.ID
	Metadata      map[string]string
	Data          Evt
	Raw           []byte
}

func (e Envelope[Evt]) Identity() SeqKey {
	return e.Key
}

type StoreOptions struct {
	Metadata  map[string]string      // information from the infrastructure e.g. correlationId, version, flags, fine to be nil
	EventTime xtime.UnixMilliseconds // an alternative domain time when the domain event was raised. If zero, the current time is used.
	CreatedBy user.ID                // the user which caused the domain event. If zero, the current subject is used.
}

// Store just persists the given Event technically which is usually nothing which a domain expert cares about.
// However, this creates a generic foundation to just persist an event at the end, e.g. if a use cases emitted a fact.
// The Evt type represents an interface sum type.
type Store[Evt any] func(subject auth.Subject, evt Evt, opts StoreOptions) (Envelope[Evt], error)

// Load is the technical counterpart of Store which is usually nothing a domain expert cares about.
type Load[Evt any] func(subject auth.Subject, id SeqID) (option.Opt[Envelope[Evt]], error)

// Truncate removes all events whose sequence ids are larger than the given max. Audit event logs must never be
// truncated, however for implementations which allows undo/redo must also use truncate to discard the history
// from the point of inserting a new event (forking the history).
type Truncate func(subject auth.Subject, max SeqID) error

// Replay is like load, but returns an event stream.
// All other use cases will use this under the hood to create views of the current state or history based on
// the collected facts.
// The Evt type represents an interface sum type.
type Replay[Evt any] func(subject auth.Subject, fromInc, toInc SeqID) iter.Seq2[Envelope[Evt], error]

type ReplayOptions struct {
	FromInc SeqID
	ToInc   SeqID
}

type ReplayWithIndex[Primary ~string, Evt any] func(subject auth.Subject, primary Primary, apply func(Envelope[Evt]) error, opts ReplayOptions) error

// OffsetsForTimestamps returns all those sequence identifiers whose event timestamps fall into the given interval.
// The system keeps an inverted index of all timestamps pointing to (multiple) sequence ids.
type OffsetsForTimestamps func(subject auth.Subject, fromInc, toInc xtime.UnixMilliseconds) iter.Seq2[SeqID, error]

// Register associates the given Event type and the given name for json marshaling and unmarshalling. You must essentially never modify
// any (union) member of Evt in an incompatible way. Adding optional fields is always fine, however renaming field names or changing underlying field types will
// cause silent data loss or errors when loaded at runtime.
type Register[Evt any] func(t reflect.Type, discriminatorName Discriminator) error

// ReadAll loops over all available sequence ids.
type ReadAll[Evt any] func(subject auth.Subject) iter.Seq2[SeqKey, error]

// MakeType creates a new event instance based on the given discriminator.
type MakeType[Evt any] func(discriminator Discriminator) (Evt, error)

// Delete removes the event with the given sequence identifier. This can be dangerous or unwanted for auditable
// setups. Usually this is only useful for debugging, testing, or repairs but may also be used to truncate the history
// to avoid unbound data grow or to comply with GDPR requests.
type Delete[Evt any] func(subject auth.Subject, id SeqID) error

// DeleteByPrimary removes all events with the given primary key.
type DeleteByPrimary[Evt any] func(subject auth.Subject, index IdxID, primary string) error

type RegisteredType struct {
	Type          reflect.Type
	Discriminator Discriminator
}

// RegisteredTypes returns all registered event types.
type RegisteredTypes[Evt any] func() iter.Seq[RegisteredType]

type Options[Evt any] struct {
	// Mutex is optional and may be nil. If nil, a new mutex is created automatically to protect critical section
	// under currency pressure. Otherwise, updates and creates may accidentally recreate or overwrite entities.
	Mutex *sync.Mutex

	// Bus may be nil. If not, according mutation events are eventually published.
	Bus events.Bus

	// Indexer is evaluated on insertion into the event store and will return those strings which
	// are used as a composite key lookup.
	Indexer []Indexer[Evt]
}

type UseCases[Evt any] struct {
	Store                Store[Evt]
	Load                 Load[Evt]
	Truncate             Truncate
	Replay               Replay[Evt]
	OffsetsForTimestamps OffsetsForTimestamps
	Register             Register[Evt]
	ReadAll              ReadAll[Evt]
	RegisteredTypes      RegisteredTypes[Evt]
	MakeType             MakeType[Evt]
	Delete               Delete[Evt]
	DeleteByPrimary      DeleteByPrimary[Evt]
}

func NewUseCases[Evt any](perms Permissions, eventStore blob.Store, timeStore blob.Store, opts Options[Evt]) UseCases[Evt] {
	var typeRegistry concurrent.RWMap[reflect.Type, Discriminator]
	var invTypeRegistry concurrent.RWMap[Discriminator, reflect.Type]

	loadFn := NewLoad[Evt](perms, eventStore, &invTypeRegistry)
	deleteFn := NewDelete(perms, opts.Mutex, loadFn, eventStore, timeStore, opts)

	return UseCases[Evt]{
		Store:                NewStore[Evt](perms, opts.Mutex, &typeRegistry, eventStore, timeStore, opts),
		Load:                 loadFn,
		Truncate:             nil,
		Replay:               NewReplay[Evt](perms, eventStore, loadFn),
		OffsetsForTimestamps: nil,
		Register:             NewRegister[Evt](&typeRegistry, &invTypeRegistry),
		ReadAll:              NewReadAll[Evt](perms, eventStore),
		RegisteredTypes:      NewRegisteredTypes[Evt](&invTypeRegistry),
		MakeType:             NewMakeType[Evt](&invTypeRegistry),
		Delete:               deleteFn,
		DeleteByPrimary:      NewDeleteByPrimary(perms, deleteFn, opts),
	}
}
