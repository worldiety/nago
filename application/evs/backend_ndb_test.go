// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs_test

import (
	"context"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

// openNDBMessages opens a fresh msgstore-backed ndb.Messages in a temp dir.
func openNDBMessages(t *testing.T) ndb.Messages {
	t.Helper()
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: msgstore.EngineKind}))
	t.Cleanup(func() { option.MustZero(db.Close()) })

	eng, err := db.Engine("events", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	me, ok := eng.(ndb.MessageEngine)
	if !ok {
		t.Fatal("expected a message engine")
	}
	return me.Messages()
}

// TestHandlerOverNDBBackend exercises the exact same Handler/decide-evolve flow
// as TestHandler, but on the real ndb message store via NewNDBBackend, proving
// the two backends are interchangeable behind the Backend seam.
func TestHandlerOverNDBBackend(t *testing.T) {
	msgs := openNDBMessages(t)
	backend := evs.NewNDBBackend[Evt, *Person](msgs)

	aggID := func(e Evt) (PID, bool) {
		switch evt := e.(type) {
		case FirstnameUpdated:
			return evt.Person, evt.Person != ""
		case LastnameUpdated:
			return evt.Person, evt.Person != ""
		default:
			return "", false
		}
	}

	handler := evs.NewHandler[*Person](backend, aggID, backend.Register)
	handler.RegisterEvents(FirstnameUpdated{}, LastnameUpdated{})

	if err := handler.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "John"}); err != nil {
		t.Fatalf("handle: %v", err)
	}

	got := option.Must(handler.Aggregate(context.Background(), "1"))
	if got.firstname != "John" {
		t.Fatalf("firstname: got %q want John", got.firstname)
	}

	// A second aggregate must stay independent.
	if err := handler.Handle(user.SU(), "2", UpdateFirstnameCmd{ID: "2", Firstname: "Jane"}); err != nil {
		t.Fatalf("handle 2: %v", err)
	}
	if option.Must(handler.Aggregate(context.Background(), "2")).firstname != "Jane" {
		t.Fatal("second aggregate firstname wrong")
	}
	if option.Must(handler.Aggregate(context.Background(), "1")).firstname != "John" {
		t.Fatal("first aggregate corrupted by second")
	}
}

// TestHandlerNDBReplayRebuild verifies that a fresh handler over the same store
// rebuilds state purely from the persisted events (eager replay).
func TestHandlerNDBReplayRebuild(t *testing.T) {
	msgs := openNDBMessages(t)

	aggID := func(e Evt) (PID, bool) {
		switch evt := e.(type) {
		case FirstnameUpdated:
			return evt.Person, evt.Person != ""
		case LastnameUpdated:
			return evt.Person, evt.Person != ""
		default:
			return "", false
		}
	}

	// first handler writes some events
	{
		backend := evs.NewNDBBackend[Evt, *Person](msgs)
		h := evs.NewHandler[*Person](backend, aggID, backend.Register)
		h.RegisterEvents(FirstnameUpdated{}, LastnameUpdated{})
		if err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "John"}); err != nil {
			t.Fatalf("handle: %v", err)
		}
	}

	// second handler over the SAME messages must see the state via replay
	backend2 := evs.NewNDBBackend[Evt, *Person](msgs)
	h2 := evs.NewHandler[*Person](backend2, aggID, backend2.Register)
	h2.RegisterEvents(FirstnameUpdated{}, LastnameUpdated{})

	got := option.Must(h2.Aggregate(context.Background(), "1"))
	if got.firstname != "John" {
		t.Fatalf("rebuilt firstname: got %q want John", got.firstname)
	}
}
