// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package msgstore_test

import (
	"math"
	"sync"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

// TestConcurrentAppendSameTypeIsOrdered guards the invariant that the messages
// of a single type are laid out on disk in ascending Seq order.
//
// Append documents that concurrent appends to the same type are supported and
// serialized by the per-type mutex. The Seq must therefore be allocated while
// that mutex is held: if it were allocated before, two writers could acquire
// their sequence numbers in one order and win the mutex in the opposite one,
// writing a lower Seq behind a higher one. That breaks three things at once —
// replayType's early return on maxSeq, the ascending-cursor precondition of the
// k-way merge heap, and the segment's finalized maxSeq (which is derived from
// lastSeq and encoded into the file name, so the damage survives a reopen).
func TestConcurrentAppendSameTypeIsOrdered(t *testing.T) {
	const eventType msgstore.TypeID = "concurrent"
	const writers = 8
	const perWriter = 64

	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	traceID := msgstore.NewTraceID()
	payload := []byte("payload")

	// force the type state to exist before the racing writers start, so that
	// lazy initialisation is not what is being tested here
	option.Must(db.Append(eventType, traceID, payload))

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(writers)
	for range writers {
		go func() {
			defer wg.Done()
			<-start
			for range perWriter {
				if _, err := db.Append(eventType, traceID, payload); err != nil {
					t.Errorf("append: %v", err)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()

	// the on-disk order of a single type must be strictly ascending, and every
	// appended message must still be reachable
	var seqs []msgstore.Seq
	for _, msg := range db.Replay(nil, 0, msgstore.Seq(math.MaxUint64)) {
		seqs = append(seqs, msg.Seq)
	}

	want := 1 + writers*perWriter
	if len(seqs) != want {
		t.Errorf("replayed %d messages, want %d", len(seqs), want)
	}
	for i := 1; i < len(seqs); i++ {
		if seqs[i] <= seqs[i-1] {
			t.Fatalf("replay not strictly ascending at index %d: %d after %d", i, seqs[i], seqs[i-1])
		}
	}
}

// TestConcurrentAppendSameTypeSurvivesReopen is the persistent half of
// TestConcurrentAppendSameTypeIsOrdered: an out-of-order write corrupts the
// finalized segment's maxSeq, which is encoded into the file name and therefore
// makes shouldSkipSegment drop the segment on a later bounded replay. Splitting
// eagerly forces the finalize path to run.
func TestConcurrentAppendSameTypeSurvivesReopen(t *testing.T) {
	const eventType msgstore.TypeID = "concurrent"
	const writers = 8
	const perWriter = 64

	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
		// split often so that segments get finalized (and named) during the run
		ShouldSplit: func(info msgstore.SegmentInfo) bool { return info.MessageCount >= 16 },
	}))

	traceID := msgstore.NewTraceID()
	payload := []byte("payload")
	option.Must(db.Append(eventType, traceID, payload))

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(writers)
	for range writers {
		go func() {
			defer wg.Done()
			<-start
			for range perWriter {
				if _, err := db.Append(eventType, traceID, payload); err != nil {
					t.Errorf("append: %v", err)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	option.MustZero(db.Close())

	db2 := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db2.Close()) }()

	var seqs []msgstore.Seq
	for _, msg := range db2.Replay(nil, 0, msgstore.Seq(math.MaxUint64)) {
		seqs = append(seqs, msg.Seq)
	}

	want := 1 + writers*perWriter
	if len(seqs) != want {
		t.Errorf("after reopen replayed %d messages, want %d", len(seqs), want)
	}
	for i := 1; i < len(seqs); i++ {
		if seqs[i] <= seqs[i-1] {
			t.Fatalf("after reopen not strictly ascending at index %d: %d after %d", i, seqs[i], seqs[i-1])
		}
	}
}
