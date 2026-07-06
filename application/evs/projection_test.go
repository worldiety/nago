// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs_test

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/ndb"
)

// reflectTypeOf is a tiny readability shim so the many backend.Register lines
// read as intent ("register this event value") rather than reflection noise.
func reflectTypeOf(v any) reflect.Type { return reflect.TypeOf(v) }

// personView is a flat read-model row folded from the person events. It is a
// value type with no mutable reference fields, so Projection.Get can hand out a
// safe shallow copy.
type personView struct {
	First string
	Last  string
	// Events counts how many events were folded into this row (proves ordering
	// and that both event types reach the same projection under one key).
	Events int
}

// waitForProcessed spins (cheaply) until the projection has folded up to seq or
// the deadline passes. Used where a test wrote via the store directly and wants
// the tail goroutine to catch up.
func waitProcessed(t *testing.T, p interface{ ProcessedSeq() ndb.Seq }, seq ndb.Seq, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if p.ProcessedSeq() >= seq {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("projection did not reach seq %d in %s (at %d)", seq, d, p.ProcessedSeq())
}

// appendPerson writes one event through the real ndb backend and returns its
// assigned global sequence.
func appendPerson(t *testing.T, backend evs.RegisterableBackend[Evt], e Evt) ndb.Seq {
	t.Helper()
	env, err := backend.Append(user.SU(), e)
	if err != nil {
		t.Fatalf("append %T: %v", e, err)
	}
	return ndb.Seq(env.Sequence)
}

func newPersonDetail(src evs.Source) *evs.Projection[PID, personView] {
	detail := evs.NewProjection[PID, personView](src, evs.ProjectionOptions{})
	evs.Project(detail,
		func(e FirstnameUpdated) PID { return e.Person },
		func(s *personView, e FirstnameUpdated) { s.First = e.Firstname; s.Events++ },
	)
	evs.Project(detail,
		func(e LastnameUpdated) PID { return e.Person },
		func(s *personView, e LastnameUpdated) { s.Last = e.Name; s.Events++ },
	)
	return detail
}

// TestProjectionWarmupThenLive proves the same code path serves history
// (events written before Run) and live updates (events written after Run), in
// order, into a resident map.
func TestProjectionWarmupThenLive(t *testing.T) {
	msgs := openNDBMessages(t)
	backend := evs.NewNDBBackend[Evt, *Person](msgs)
	option.MustZero(backend.Register(reflectTypeOf(FirstnameUpdated{}), "FirstnameUpdated"))
	option.MustZero(backend.Register(reflectTypeOf(LastnameUpdated{}), "LastnameUpdated"))

	// history: two events before the projection runs
	appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: "John"})
	lastHist := appendPerson(t, backend, LastnameUpdated{Person: "1", Name: "Doe"})

	detail := newPersonDetail(msgs)
	stop := detail.Run()
	defer stop()

	waitProcessed(t, detail, lastHist, 2*time.Second)

	got, ok := detail.Get("1")
	if !ok {
		t.Fatal("expected person 1 after warm-up")
	}
	if got.First != "John" || got.Last != "Doe" || got.Events != 2 {
		t.Fatalf("warm-up fold wrong: %+v", got)
	}

	// live: a further event must be picked up without re-running
	live := appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: "Jonathan"})
	waitProcessed(t, detail, live, 2*time.Second)

	got, _ = detail.Get("1")
	if got.First != "Jonathan" || got.Events != 3 {
		t.Fatalf("live fold wrong: %+v", got)
	}
}

// TestProjectionCrossKey is the core motivation: the SAME event type folds into
// two different projections under two different keys. Here one projection keys
// by person id and a second keys by (a derived) constant bucket, proving a rule
// chooses its own key independent of any single aggregate identity.
func TestProjectionCrossKey(t *testing.T) {
	msgs := openNDBMessages(t)
	backend := evs.NewNDBBackend[Evt, *Person](msgs)
	option.MustZero(backend.Register(reflectTypeOf(FirstnameUpdated{}), "FirstnameUpdated"))

	// projection A: keyed by person id
	byPerson := evs.NewProjection[PID, personView](msgs, evs.ProjectionOptions{})
	evs.Project(byPerson,
		func(e FirstnameUpdated) PID { return e.Person },
		func(s *personView, e FirstnameUpdated) { s.First = e.Firstname; s.Events++ },
	)

	// projection B: a global counter over the very same event type (singleton),
	// i.e. a different target type AND a different key from the same event.
	total := evs.NewSingleton[personView](msgs, evs.ProjectionOptions{})
	evs.Project(total,
		func(e FirstnameUpdated) evs.Unit { return evs.TheUnit() },
		func(s *personView, e FirstnameUpdated) { s.Events++ },
	)

	stopA := byPerson.Run()
	defer stopA()
	stopB := total.Run()
	defer stopB()

	appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: "A"})
	appendPerson(t, backend, FirstnameUpdated{Person: "2", Firstname: "B"})
	last := appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: "C"})

	waitProcessed(t, byPerson, last, 2*time.Second)
	waitProcessed(t, total, last, 2*time.Second)

	p1, _ := byPerson.Get("1")
	p2, _ := byPerson.Get("2")
	if p1.First != "C" || p1.Events != 2 {
		t.Fatalf("person 1 wrong: %+v", p1)
	}
	if p2.First != "B" || p2.Events != 1 {
		t.Fatalf("person 2 wrong: %+v", p2)
	}

	sum, ok := evs.Value(total)
	if !ok || sum.Events != 3 {
		t.Fatalf("singleton total wrong: %+v ok=%v", sum, ok)
	}
}

// TestProjectionReadYourWrite proves WaitFor(seq) makes a write observable: after
// it returns, Get reflects the event whose Seq was awaited.
func TestProjectionReadYourWrite(t *testing.T) {
	msgs := openNDBMessages(t)
	backend := evs.NewNDBBackend[Evt, *Person](msgs)
	option.MustZero(backend.Register(reflectTypeOf(FirstnameUpdated{}), "FirstnameUpdated"))

	detail := newPersonDetail(msgs)
	stop := detail.Run()
	defer stop()

	// hammer: write then immediately wait+read many times; must always see it.
	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("n%d", i)
		seq := appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: name})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := detail.WaitFor(ctx, seq); err != nil {
			cancel()
			t.Fatalf("WaitFor(%d): %v", seq, err)
		}
		cancel()

		got, ok := detail.Get("1")
		if !ok || got.First != name {
			t.Fatalf("read-your-write violated at i=%d: got %+v ok=%v", i, got, ok)
		}
	}
}

// TestProjectionWaitForContextCancel verifies WaitFor honours context
// cancellation when the target seq is never reached.
func TestProjectionWaitForContextCancel(t *testing.T) {
	msgs := openNDBMessages(t)
	detail := newPersonDetail(msgs)
	stop := detail.Run()
	defer stop()

	// nothing was ever written, so seq 999 is unreachable; ctx must win.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := detail.WaitFor(ctx, 999)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

// TestProjectionToleratesBadFold proves a panicking fold does not kill the tail
// goroutine: the bad event is skipped (via OnError) and later events still fold.
func TestProjectionToleratesBadFold(t *testing.T) {
	msgs := openNDBMessages(t)
	backend := evs.NewNDBBackend[Evt, *Person](msgs)
	option.MustZero(backend.Register(reflectTypeOf(FirstnameUpdated{}), "FirstnameUpdated"))

	var errCount atomic.Int64
	detail := evs.NewProjection[PID, personView](msgs, evs.ProjectionOptions{
		OnError: func(seq ndb.Seq, typeID ndb.TypeID, err error) { errCount.Add(1) },
	})
	evs.Project(detail,
		func(e FirstnameUpdated) PID { return e.Person },
		func(s *personView, e FirstnameUpdated) {
			// deliberately panic on a poison value to exercise recover
			if e.Firstname == "BOOM" {
				panic("boom")
			}
			s.First = e.Firstname
			s.Events++
		},
	)
	stop := detail.Run()
	defer stop()

	appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: "ok1"})
	appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: "BOOM"})
	last := appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: "ok2"})

	waitProcessed(t, detail, last, 2*time.Second)

	if errCount.Load() != 1 {
		t.Fatalf("expected exactly one OnError, got %d", errCount.Load())
	}
	got, _ := detail.Get("1")
	// the panicking event was skipped, so only ok1 and ok2 counted; last write wins.
	if got.First != "ok2" || got.Events != 2 {
		t.Fatalf("projection did not recover cleanly: %+v", got)
	}
}

// TestProjectionRunIdempotent verifies Run starts at most one goroutine and
// returns a stable stop function.
func TestProjectionRunIdempotent(t *testing.T) {
	msgs := openNDBMessages(t)
	detail := newPersonDetail(msgs)

	stop1 := detail.Run()
	stop2 := detail.Run()

	// same function value → no second goroutine was started
	if fmt.Sprintf("%p", stop1) != fmt.Sprintf("%p", stop2) {
		t.Fatal("Run should return the same stop function on repeated calls")
	}
	stop1()
	stop2() // second stop must be a harmless no-op
}

// TestProjectionAllSnapshot verifies All returns every folded row.
func TestProjectionAllSnapshot(t *testing.T) {
	msgs := openNDBMessages(t)
	backend := evs.NewNDBBackend[Evt, *Person](msgs)
	option.MustZero(backend.Register(reflectTypeOf(FirstnameUpdated{}), "FirstnameUpdated"))

	detail := newPersonDetail(msgs)
	stop := detail.Run()
	defer stop()

	appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: "A"})
	appendPerson(t, backend, FirstnameUpdated{Person: "2", Firstname: "B"})
	last := appendPerson(t, backend, FirstnameUpdated{Person: "3", Firstname: "C"})
	waitProcessed(t, detail, last, 2*time.Second)

	seen := map[PID]string{}
	for k, v := range detail.All() {
		seen[k] = v.First
	}
	if len(seen) != 3 || seen["1"] != "A" || seen["2"] != "B" || seen["3"] != "C" {
		t.Fatalf("All snapshot wrong: %+v", seen)
	}
}

// TestProjectionConcurrentReadWrite hammers concurrent reads against a running
// tail goroutine. Under -race it proves reads never touch memory the fold
// mutates concurrently.
func TestProjectionConcurrentReadWrite(t *testing.T) {
	msgs := openNDBMessages(t)
	backend := evs.NewNDBBackend[Evt, *Person](msgs)
	option.MustZero(backend.Register(reflectTypeOf(FirstnameUpdated{}), "FirstnameUpdated"))

	detail := newPersonDetail(msgs)
	stop := detail.Run()
	defer stop()

	var wg sync.WaitGroup
	done := make(chan struct{})

	// readers
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				_, _ = detail.Get("1")
				for range detail.All() {
				}
			}
		}()
	}

	// writer
	var last ndb.Seq
	for i := 0; i < 300; i++ {
		last = appendPerson(t, backend, FirstnameUpdated{Person: "1", Firstname: fmt.Sprintf("n%d", i)})
	}
	waitProcessed(t, detail, last, 5*time.Second)
	close(done)
	wg.Wait()

	got, _ := detail.Get("1")
	if got.First != "n299" {
		t.Fatalf("final fold wrong: %+v", got)
	}
}

// TestProjectionHandleWaitForGet is the end-to-end read-your-write scenario the
// whole design targets: a command is run through the Handler (which persists to
// the same ndb store the projection tails), the returned commit seq is awaited
// via WaitFor, and the projection is then guaranteed to reflect the command.
func TestProjectionHandleWaitForGet(t *testing.T) {
	msgs := openNDBMessages(t)

	// write side: a real Handler over the ndb backend
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

	// read side: a projection tailing the very same store
	detail := newPersonDetail(msgs)
	stop := detail.Run()
	defer stop()

	for i := 0; i < 25; i++ {
		name := fmt.Sprintf("n%d", i)
		seq, err := handler.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: name})
		if err != nil {
			t.Fatalf("handle i=%d: %v", i, err)
		}
		if seq == 0 {
			t.Fatalf("expected non-zero commit seq at i=%d", i)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := detail.WaitFor(ctx, ndb.Seq(seq)); err != nil {
			cancel()
			t.Fatalf("WaitFor(%d): %v", seq, err)
		}
		cancel()

		got, ok := detail.Get("1")
		if !ok || got.First != name {
			t.Fatalf("read-your-write via Handle violated at i=%d: got %+v ok=%v", i, got, ok)
		}
	}
}
