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

// Envelope is the read-side view of a stored event. EventTime uses explicit unix
// milliseconds to avoid timezone ambiguity.
//
// EventTime, Sequence and Key are read-side metadata filled by the backend from
// the store (append time and assigned sequence); they are not supplied by the
// writer. Any provenance such as "who caused this" is the domain's concern and,
// if needed, lives inside the event payload itself (the audit UI reads a
// conventional "createdBy" field from Raw heuristically).
type Envelope[Evt any] struct {
	Key           SeqKey
	Sequence      SeqID
	Discriminator Discriminator
	EventTime     xtime.UnixMilliseconds
	Data          Evt
	// Raw is the event's own JSON encoding (not any wrapper envelope).
	Raw []byte
}

func (e Envelope[Evt]) Identity() SeqKey {
	return e.Key
}

// Store just persists the given Event technically which is usually nothing which a domain expert cares about.
// However, this creates a generic foundation to just persist an event at the end, e.g. if a use cases emitted a fact.
// The Evt type represents an interface sum type.
type Store[Evt any] func(subject auth.Subject, evt Evt) (Envelope[Evt], error)

// Load is the technical counterpart of Store which is usually nothing a domain expert cares about.
type Load[Evt any] func(subject auth.Subject, id SeqID) (option.Opt[Envelope[Evt]], error)

// Replay is like load, but returns an event stream.
// All other use cases will use this under the hood to create views of the current state or history based on
// the collected facts.
// The Evt type represents an interface sum type.
type Replay[Evt any] func(subject auth.Subject, fromInc, toInc SeqID) iter.Seq2[Envelope[Evt], error]

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
}

type UseCases[Evt any] struct {
	Store           Store[Evt]
	Load            Load[Evt]
	Replay          Replay[Evt]
	Register        Register[Evt]
	ReadAll         ReadAll[Evt]
	RegisteredTypes RegisteredTypes[Evt]
	MakeType        MakeType[Evt]
	Delete          Delete[Evt]
}

func NewUseCases[Evt any](perms Permissions, eventStore blob.Store, opts Options[Evt]) UseCases[Evt] {
	var typeRegistry concurrent.RWMap[reflect.Type, Discriminator]
	var invTypeRegistry concurrent.RWMap[Discriminator, reflect.Type]

	if opts.Mutex == nil {
		opts.Mutex = &sync.Mutex{}
	}

	loadFn := NewLoad[Evt](perms, eventStore, &invTypeRegistry)
	deleteFn := NewDelete(perms, loadFn, eventStore)

	return UseCases[Evt]{
		Store:           NewStore[Evt](perms, &typeRegistry, eventStore),
		Load:            loadFn,
		Replay:          NewReplay[Evt](perms, eventStore, loadFn),
		Register:        NewRegister[Evt](&typeRegistry, &invTypeRegistry),
		ReadAll:         NewReadAll[Evt](perms, eventStore),
		RegisteredTypes: NewRegisteredTypes[Evt](&invTypeRegistry),
		MakeType:        NewMakeType[Evt](&invTypeRegistry),
		Delete:          deleteFn,
	}
}
