// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs_test

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"sync"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
)

type PID string
type Person struct {
	id        PID
	firstname string
	lastname  string
	deleted   bool
}

func (p *Person) Clone() *Person {
	return &Person{id: p.id, firstname: p.firstname, lastname: p.lastname, deleted: p.deleted}
}

func (p *Person) IsDeleted() bool {
	return p.deleted
}

type UpdateFirstnameCmd struct {
	ID        PID
	Firstname string
}

func (cmd UpdateFirstnameCmd) Decide(subject auth.Subject, aggregate *Person) ([]Evt, error) {
	if cmd.Firstname == aggregate.firstname {
		return nil, fmt.Errorf("firstname did not change")
	}

	if cmd.Firstname == "" {
		return nil, fmt.Errorf("firstname cannot be empty")
	}

	return []Evt{FirstnameUpdated{Person: cmd.ID, Firstname: cmd.Firstname}}, nil
}

type Evt = evs.Evt[*Person]

type FirstnameUpdated struct {
	Person    PID    `json:"person"`
	Firstname string `json:"firstname"`
}

func (f FirstnameUpdated) Discriminator() evs.Discriminator {
	return "FirstnameUpdated"
}

func (f FirstnameUpdated) Evolve(ctx context.Context, mut *Person) error {
	mut.firstname = f.Firstname
	return nil
}

type LastnameUpdated struct {
	Person PID    `json:"person"`
	Name   string `json:"name"`
}

func (f LastnameUpdated) Discriminator() evs.Discriminator {
	return "LastnameUpdated"
}

func (f LastnameUpdated) Evolve(ctx context.Context, mut *Person) error {
	mut.lastname = f.Name
	return nil
}

// memBackend is a minimal in-memory [evs.Backend] test double. Append assigns a
// monotonic sequence and stores the envelope; ReplayAll yields them in order.
type memBackend struct {
	events []evs.Envelope[Evt]
}

func (b *memBackend) Append(subject auth.Subject, e Evt) (evs.Envelope[Evt], error) {
	env := evs.Envelope[Evt]{
		Sequence: evs.SeqID(len(b.events) + 1),
		Data:     e,
	}
	b.events = append(b.events, env)
	return env, nil
}

func (b *memBackend) ReplayAll(subject auth.Subject) iter.Seq2[evs.Envelope[Evt], error] {
	return func(yield func(evs.Envelope[Evt], error) bool) {
		for _, env := range b.events {
			if !yield(env, nil) {
				return
			}
		}
	}
}

func TestHandler(t *testing.T) {
	backend := &memBackend{}
	handler := evs.NewHandler[*Person](backend, func(e Evt) (PID, bool) {
		switch evt := e.(type) {
		case FirstnameUpdated:
			return evt.Person, evt.Person != ""
		case LastnameUpdated:
			return evt.Person, evt.Person != ""
		default:
			return "", false
		}
	}, nil)

	handler.RegisterEvents(
		FirstnameUpdated{},
		LastnameUpdated{},
	)

	if _, err := handler.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "John"}); err != nil {
		t.Error(err)
	}

	if option.Must(handler.Aggregate(context.Background(), "1")).firstname != "John" {
		t.Error("firstname did not change")
	}

	t.Log(handler.Aggregate(context.Background(), "1"))
}

func newPersonHandler() *evs.Handler[*Person, Evt, PID] {
	backend := &memBackend{}
	h := evs.NewHandler[*Person](backend, func(e Evt) (PID, bool) {
		switch evt := e.(type) {
		case FirstnameUpdated:
			return evt.Person, evt.Person != ""
		case LastnameUpdated:
			return evt.Person, evt.Person != ""
		default:
			return "", false
		}
	}, nil)
	h.RegisterEvents(FirstnameUpdated{}, LastnameUpdated{})
	return h
}

// TestAggregateSnapshotIsolation verifies that mutating a returned aggregate
// (the reader snapshot) cannot corrupt the handler's source of truth: a later
// read reflects only what was persisted and evolved, not the stray mutation.
func TestAggregateSnapshotIsolation(t *testing.T) {
	h := newPersonHandler()
	if _, err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "John"}); err != nil {
		t.Fatal(err)
	}

	// reader accidentally mutates the returned snapshot
	snap := option.Must(h.Aggregate(context.Background(), "1"))
	snap.firstname = "CORRUPTED"

	// a subsequent legit change + read must not carry the stray mutation
	if _, err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "Jane"}); err != nil {
		t.Fatal(err)
	}
	got := option.Must(h.Aggregate(context.Background(), "1"))
	if got.firstname != "Jane" {
		t.Fatalf("truth corrupted by reader mutation: got %q want Jane", got.firstname)
	}
}

// TestAggregateSnapshotReuse verifies that repeated reads without an intervening
// write return the identical cached snapshot (no re-clone), while a write makes
// the next read observe the new value.
func TestAggregateSnapshotReuse(t *testing.T) {
	h := newPersonHandler()
	if _, err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "John"}); err != nil {
		t.Fatal(err)
	}

	a := option.Must(h.Aggregate(context.Background(), "1"))
	b := option.Must(h.Aggregate(context.Background(), "1"))
	if a != b {
		t.Fatal("expected the same cached snapshot pointer for repeated reads without a write")
	}

	if _, err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "Jane"}); err != nil {
		t.Fatal(err)
	}
	c := option.Must(h.Aggregate(context.Background(), "1"))
	if c == a {
		t.Fatal("expected a fresh snapshot after a write")
	}
	if c.firstname != "Jane" {
		t.Fatalf("snapshot did not refresh: got %q want Jane", c.firstname)
	}
}

// TestAggregateConcurrentReadWrite hammers concurrent Handle (writer) and
// Aggregate (readers) on the same aggregate. Under -race this proves the RCU
// design is data-race free: readers only ever touch the immutable snapshot while
// the writer mutates the separate live instance.
func TestAggregateConcurrentReadWrite(t *testing.T) {
	h := newPersonHandler()
	if _, err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "n0"}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// readers
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				ws, err := h.Aggregate(context.Background(), "1")
				if err == nil {
					_ = ws.firstname // read the snapshot
				}
			}
		}()
	}

	// writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 500; i++ {
			_, _ = h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: fmt.Sprintf("n%d", i)})
		}
		close(stop)
	}()

	wg.Wait()

	if option.Must(h.Aggregate(context.Background(), "1")).firstname != "n500" {
		t.Fatalf("final state wrong: %q", option.Must(h.Aggregate(context.Background(), "1")).firstname)
	}
}

// noopCmd decides to emit no events at all (a command without effect).
type noopCmd struct{}

func (noopCmd) Decide(subject auth.Subject, aggregate *Person) ([]Evt, error) {
	return nil, nil
}

// TestHandleReturnsCommitSeq verifies the sequence Handle returns: a monotonic,
// non-zero value for an effective command, and 0 for a no-op command (Decide
// produced no events), matching the "nothing to wait for" convention.
func TestHandleReturnsCommitSeq(t *testing.T) {
	h := newPersonHandler()

	seq1, err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "John"})
	if err != nil {
		t.Fatal(err)
	}
	if seq1 == 0 {
		t.Fatal("expected a non-zero commit seq for an effective command")
	}

	seq2, err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "Jane"})
	if err != nil {
		t.Fatal(err)
	}
	if seq2 <= seq1 {
		t.Fatalf("commit seq must increase: got %d after %d", seq2, seq1)
	}

	// a no-op command persists nothing → seq 0 ("nothing to wait for")
	noopSeq, err := h.Handle(user.SU(), "1", noopCmd{})
	if err != nil {
		t.Fatal(err)
	}
	if noopSeq != 0 {
		t.Fatalf("expected seq 0 for a no-op command, got %d", noopSeq)
	}
}

// TestHandleErrorReturnsZeroSeq verifies a failed decide yields seq 0.
func TestHandleErrorReturnsZeroSeq(t *testing.T) {
	h := newPersonHandler()
	// same firstname as the zero-value aggregate ("") triggers a decide error
	// only when non-empty; use an empty firstname which UpdateFirstnameCmd rejects.
	seq, err := h.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: ""})
	if err == nil {
		t.Fatal("expected an error for empty firstname")
	}
	if seq != 0 {
		t.Fatalf("expected seq 0 on error, got %d", seq)
	}
}

// TestHandleEmptyKeyIsRenderableError verifies that an empty aggregate key is
// rejected as a renderable, sentinel error (not a raw fmt.Errorf that a UI would
// hide behind an anonymous support token): it must match errors.Is(ErrEmptyKey),
// carry a localized title/description, return seq 0, and never touch the backend.
func TestHandleEmptyKeyIsRenderableError(t *testing.T) {
	backend := &memBackend{}
	h := evs.NewHandler[*Person](backend, func(e Evt) (PID, bool) {
		switch evt := e.(type) {
		case FirstnameUpdated:
			return evt.Person, evt.Person != ""
		default:
			return "", false
		}
	}, nil)
	h.RegisterEvents(FirstnameUpdated{}, LastnameUpdated{})

	seq, err := h.Handle(user.SU(), "", UpdateFirstnameCmd{ID: "", Firstname: "John"})
	if err == nil {
		t.Fatal("expected an error for an empty key")
	}
	if seq != 0 {
		t.Fatalf("expected seq 0 on error, got %d", seq)
	}
	if !errors.Is(err, evs.ErrEmptyKey) {
		t.Fatalf("expected errors.Is(err, ErrEmptyKey), got %v", err)
	}

	var loc std.LocalizedError
	if !errors.As(err, &loc) {
		t.Fatalf("expected a std.LocalizedError so the UI can render it, got %T", err)
	}
	if loc.Title() == "" || loc.Description() == "" {
		t.Fatalf("expected a non-empty localized title/description, got %q / %q", loc.Title(), loc.Description())
	}

	// the guard must run before any persistence: nothing may have been appended.
	if len(backend.events) != 0 {
		t.Fatalf("expected no events appended for an empty key, got %d", len(backend.events))
	}
}
