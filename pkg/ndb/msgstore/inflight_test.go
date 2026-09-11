// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package msgstore

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestInflightWatermarkHoldsBackPendingSeq is the core property: the watermark
// must never cover a sequence number whose write has not resolved, even when a
// later one resolved first. Reporting too high is what makes a follower mistake
// a write in progress for a permanent hole and drop it.
func TestInflightWatermarkHoldsBackPendingSeq(t *testing.T) {
	var counter atomic.Uint64
	counter.Store(1)
	var s inflightSet

	if got := s.committed(); got != 0 {
		t.Fatalf("empty set: committed = %d, want 0", got)
	}

	a := s.allocate(&counter) // 1
	b := s.allocate(&counter) // 2
	c := s.allocate(&counter) // 3

	if got := s.committed(); got != 0 {
		t.Fatalf("all pending: committed = %d, want 0", got)
	}

	// resolve out of order: the watermark must stay behind the oldest pending
	s.end(c)
	if got := s.committed(); got != 0 {
		t.Fatalf("after ending %d: committed = %d, want 0 (%d still pending)", c, got, a)
	}
	s.end(b)
	if got := s.committed(); got != 0 {
		t.Fatalf("after ending %d: committed = %d, want 0 (%d still pending)", b, got, a)
	}

	// only once the oldest resolves may the watermark jump to the top
	s.end(a)
	if got := s.committed(); got != c {
		t.Fatalf("all resolved: committed = %d, want %d", got, c)
	}
}

// TestInflightWatermarkIsMonotonic guards the contract that CommittedSeq never
// moves backwards, including across the transition from a non-empty to an empty
// set (where the value switches from "oldest pending minus one" to "highest
// retired").
func TestInflightWatermarkIsMonotonic(t *testing.T) {
	var counter atomic.Uint64
	counter.Store(1)
	var s inflightSet

	var last uint64
	check := func(where string) {
		got := s.committed()
		if got < last {
			t.Fatalf("%s: committed went backwards: %d after %d", where, got, last)
		}
		last = got
	}

	for round := range 20 {
		var pending []uint64
		for range 4 {
			pending = append(pending, s.allocate(&counter))
			check("after allocate")
		}
		// retire in a rotating order so the oldest is not always released first
		for i := range pending {
			s.end(pending[(i+round)%len(pending)])
			check("after end")
		}
	}
}

// TestInflightAllocateIsAtomic checks that allocation and registration cannot be
// observed apart. If they could, a concurrent reader could see an empty set
// while a sequence number had already been handed out, and the watermark would
// briefly cover a write that has not even started.
func TestInflightAllocateIsAtomic(t *testing.T) {
	var counter atomic.Uint64
	counter.Store(1)
	var s inflightSet

	const writers = 16
	const perWriter = 200

	var stop atomic.Bool
	var observerDone sync.WaitGroup
	observerDone.Add(1)
	var violation atomic.Uint64
	go func() {
		defer observerDone.Done()
		var prev uint64
		for !stop.Load() {
			got := s.committed()
			if got < prev {
				violation.Store(got)
				return
			}
			prev = got
		}
	}()

	var wg sync.WaitGroup
	wg.Add(writers)
	for range writers {
		go func() {
			defer wg.Done()
			for range perWriter {
				seq := s.allocate(&counter)
				// the watermark must never already cover a Seq we are holding
				if c := s.committed(); c >= seq {
					t.Errorf("committed %d covers still-pending seq %d", c, seq)
				}
				s.end(seq)
			}
		}()
	}
	wg.Wait()
	stop.Store(true)
	observerDone.Wait()

	if v := violation.Load(); v != 0 {
		t.Fatalf("watermark went backwards at %d", v)
	}
	if got, want := s.committed(), uint64(writers*perWriter); got != want {
		t.Fatalf("final committed = %d, want %d", got, want)
	}
}
