// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package eventstore

import (
	"sort"
	"sync"
	"testing"
	"time"
)

// TestNewIDSortsLexicographically is the regression guard for the unpadded sequence number.
// The store locates events with range scans over the raw string keys, so "10" sorting before
// "9" silently broke the replay order once ten events landed in the same millisecond.
func TestNewIDSortsLexicographically(t *testing.T) {
	const count = 5000

	ids := make([]ID, 0, count)
	for range count {
		ids = append(ids, NewID())
	}

	for i := 1; i < len(ids); i++ {
		if ids[i-1] >= ids[i] {
			t.Fatalf("ids are not strictly increasing as strings at %d: %q then %q", i, ids[i-1], ids[i])
		}
	}
}

func TestNewIDSequenceNeverExceedsSingleDigit(t *testing.T) {
	for range 5000 {
		id := NewID()

		seq := string(id[len(id)-1])
		if id[len(id)-2] != '/' {
			t.Fatalf("sequence number is not a single digit: %q", id)
		}

		if seq < "0" || seq > "9" {
			t.Fatalf("unexpected sequence number %q in %q", seq, id)
		}
	}
}

// TestNewIDConcurrentUniqueness guards the collision that a pair of atomics allowed: two
// goroutines entering a new millisecond together both reset the counter and both received
// sequence number zero, so one event overwrote the other.
func TestNewIDConcurrentUniqueness(t *testing.T) {
	const (
		workers   = 16
		perWorker = 200
	)

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		all []ID
	)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			local := make([]ID, 0, perWorker)
			for range perWorker {
				local = append(local, NewID())
			}

			mu.Lock()
			all = append(all, local...)
			mu.Unlock()
		}()
	}

	wg.Wait()

	seen := make(map[ID]bool, len(all))
	for _, id := range all {
		if seen[id] {
			t.Fatalf("duplicate id handed out: %q", id)
		}
		seen[id] = true
	}

	// the global order must still be consistent
	sorted := make([]ID, len(all))
	copy(sorted, all)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	for i := 1; i < len(sorted); i++ {
		if sorted[i-1] == sorted[i] {
			t.Fatalf("duplicate after sorting at %d: %q", i, sorted[i])
		}
	}
}

// TestReserveSeqNoWaitsWhenMillisecondIsFull pins the throttle itself, without paying for wall
// clock time.
func TestReserveSeqNoWaitsWhenMillisecondIsFull(t *testing.T) {
	idMutex.Lock()
	lastUnixMilli = 1000
	lastSeqNo = 0
	idMutex.Unlock()

	for want := int64(1); want <= maxSeqNoPerMilli; want++ {
		got, ok := reserveSeqNo(1000)
		if !ok || got != want {
			t.Fatalf("reserve %d: got %d, ok=%v", want, got, ok)
		}
	}

	if _, ok := reserveSeqNo(1000); ok {
		t.Error("an exhausted millisecond must report false so the caller waits")
	}

	// the next millisecond starts over
	if got, ok := reserveSeqNo(1001); !ok || got != 0 {
		t.Errorf("new millisecond: got %d, ok=%v", got, ok)
	}
}

// TestReserveSeqNoRejectsBackwardsClock pins that a clock step backwards makes the caller wait
// instead of emitting an id which would sort into the past.
func TestReserveSeqNoRejectsBackwardsClock(t *testing.T) {
	idMutex.Lock()
	lastUnixMilli = 2000
	lastSeqNo = 0
	idMutex.Unlock()

	if _, ok := reserveSeqNo(1999); ok {
		t.Error("an id from before the last handed out millisecond must be refused")
	}
}

func TestNewIDThroughputIsCapped(t *testing.T) {
	// 25 ids need at least two full milliseconds at ten per millisecond
	start := time.Now()
	for range 25 {
		NewID()
	}

	if d := time.Since(start); d < time.Millisecond {
		t.Errorf("expected the throttle to engage, took only %v", d)
	}
}
