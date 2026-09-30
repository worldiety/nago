// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func waitDone(t *testing.T, l *EventLoop) {
	t.Helper()

	select {
	case <-l.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("event loop did not exit")
	}
}

// TestEventLoopPostWhileDestroy posts from many goroutines while the loop is destroyed. This must neither panic
// nor race, and each accepted function must be executed exactly once.
func TestEventLoopPostWhileDestroy(t *testing.T) {
	for range 200 {
		l := NewEventLoop()

		var accepted, executed atomic.Int64
		var wg sync.WaitGroup
		for range 4 {
			wg.Go(func() {
				for range 50 {
					if l.Post(func() { executed.Add(1) }) {
						accepted.Add(1)
					}
				}
			})
		}

		l.Destroy()
		wg.Wait()
		waitDone(t, l)

		if accepted.Load() != executed.Load() {
			t.Fatalf("accepted %d but executed %d", accepted.Load(), executed.Load())
		}

		if l.Post(func() {}) {
			t.Fatal("post after destroy must be rejected")
		}

		if l.Destroy() {
			t.Fatal("second destroy must be rejected")
		}
	}
}

// TestEventLoopOrder checks that functions posted before the destruction and the final functions are executed
// in order, and that nothing runs afterward.
func TestEventLoopOrder(t *testing.T) {
	l := NewEventLoop()

	var got []int
	for i := range 10 {
		l.Post(func() { got = append(got, i) })
	}

	l.Destroy(func() { got = append(got, 10) }, func() { got = append(got, 11) })
	l.Post(func() { got = append(got, -1) })
	waitDone(t, l)

	for i, v := range got {
		if v != i {
			t.Fatalf("unexpected order %v", got)
		}
	}

	if len(got) != 12 {
		t.Fatalf("unexpected executions %v", got)
	}

	if l.Pending() != 0 {
		t.Fatalf("expected no pending functions, got %d", l.Pending())
	}
}

// TestEventLoopPostFromLoop posts more functions from within the loop than a bounded queue would accept, which
// must not deadlock.
func TestEventLoopPostFromLoop(t *testing.T) {
	l := NewEventLoop()
	defer l.Destroy()

	done := make(chan struct{})
	l.Post(func() {
		for range 10_000 {
			l.Post(func() {})
		}

		l.Post(func() { close(done) })
	})

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("event loop deadlocked")
	}
}

// TestEventLoopDestroyFromLoop destroys the loop from within itself.
func TestEventLoopDestroyFromLoop(t *testing.T) {
	l := NewEventLoop()

	var final atomic.Bool
	l.Post(func() {
		l.Destroy(func() { final.Store(true) })
	})

	waitDone(t, l)
	if !final.Load() {
		t.Fatal("final function has not been executed")
	}
}

// TestUpdateTickIsCoalesced ensures that a blocked scope accumulates at most a single tick.
func TestUpdateTickIsCoalesced(t *testing.T) {
	s := &Scope{eventLoop: NewEventLoop()}
	defer s.eventLoop.Destroy()

	release := make(chan struct{})
	s.eventLoop.Post(func() { <-release })

	for range 100 {
		s.updateTick(time.Now())
	}

	if p := s.eventLoop.Pending(); p != 2 {
		t.Fatalf("expected the blocking function and a single tick, got %d", p)
	}

	close(release)
}
