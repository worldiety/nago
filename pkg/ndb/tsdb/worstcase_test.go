// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"math/rand"
	"testing"
)

// TestWorstCaseRandomUnorderedReopen is the adversarial case for tsdb: one
// million points with fully random, unordered timestamps and random values.
//
// Because timestamps are not monotonic, essentially every write misses the
// append fast path and is routed through the out-of-order head + compaction
// machinery — the most expensive path. The data is inserted in 10 batches of
// 100k, and the database is fully closed and reopened after every batch, forcing
// head-WAL replay, pending-chunk finalization and chunk-merge compaction to be
// exercised repeatedly. Finally we reopen once more and verify that every point
// is present with its correct (newest-wins) value.
func TestWorstCaseRandomUnorderedReopen(t *testing.T) {
	const (
		total     = 1_000_000
		steps     = 10
		perStep   = total / steps
		tsSpan    = int64(5_000_000) // timestamp domain; < total so collisions occur
		valDomain = 1_000_000
	)

	dir := t.TempDir()
	rng := rand.New(rand.NewSource(0xBADC0FFEE))

	// expected[ts] = newest value written at ts (overwrite semantics).
	expected := make(map[int64]int64, total)

	for step := 0; step < steps; step++ {
		db, err := Open(dir, Options{Compress: false})
		if err != nil {
			t.Fatalf("step %d: open: %v", step, err)
		}
		c, err := db.Column("worst", "sig", Schema{Scheme: SchemeDecimal, Decimals: 0})
		if err != nil {
			t.Fatalf("step %d: column: %v", step, err)
		}

		for i := 0; i < perStep; i++ {
			ts := rng.Int63n(tsSpan)
			// avoid the NA sentinel so a stored value is never mistaken for a hole
			v := int64(rng.Intn(valDomain)) - valDomain/2
			if v == NA {
				v = 0
			}
			if err := c.PutI64(ts, v); err != nil {
				t.Fatalf("step %d: put: %v", step, err)
			}
			expected[ts] = v // newest write wins
		}

		if err := db.Close(); err != nil {
			t.Fatalf("step %d: close: %v", step, err)
		}
	}

	// Final reopen: fully independent handle, everything must be recovered.
	db, err := Open(dir, Options{Compress: false})
	if err != nil {
		t.Fatalf("final open: %v", err)
	}
	defer db.Close()

	c, ok, err := db.LookupColumn("worst", "sig")
	if err != nil {
		t.Fatalf("final lookup: %v", err)
	}
	if !ok {
		t.Fatal("final lookup: column missing after reopen")
	}

	got := make(map[int64]int64, len(expected))
	var (
		count  int64
		lastTS int64 = -1
		haveTS bool
	)
	err = c.ScanI64(minInt64, maxInt64, func(ts, vals []int64) bool {
		for i := range ts {
			// ordering invariant: strictly ascending, unique timestamps
			if haveTS && ts[i] <= lastTS {
				t.Errorf("scan not strictly ascending: ts[%d]=%d after %d", i, ts[i], lastTS)
			}
			lastTS = ts[i]
			haveTS = true
			got[ts[i]] = vals[i]
			count++
		}
		return true
	})
	if err != nil {
		t.Fatalf("final scan: %v", err)
	}

	// 1) exact cardinality: as many distinct points as distinct timestamps written
	if int(count) != len(expected) {
		t.Fatalf("point count = %d, want %d distinct timestamps", count, len(expected))
	}
	if len(got) != len(expected) {
		t.Fatalf("distinct read timestamps = %d, want %d", len(got), len(expected))
	}

	// 2) every expected point present with its newest value
	missing, wrong := 0, 0
	for ts, want := range expected {
		v, present := got[ts]
		if !present {
			if missing < 10 {
				t.Errorf("missing ts=%d (want value %d)", ts, want)
			}
			missing++
			continue
		}
		if v != want {
			if wrong < 10 {
				t.Errorf("ts=%d value=%d, want %d", ts, v, want)
			}
			wrong++
		}
	}
	if missing > 0 || wrong > 0 {
		t.Fatalf("verification failed: %d missing, %d wrong (of %d)", missing, wrong, len(expected))
	}

	// 3) no phantom points that were never written
	for ts := range got {
		if _, present := expected[ts]; !present {
			t.Fatalf("phantom ts=%d present but never written", ts)
		}
	}

	t.Logf("verified %d distinct points across %d random unordered batches with reopen", len(expected), steps)
}
