// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/ndb"
)

// TestResetFoldedKeepsWhatCameAfterTheSnapshot: a compaction folds a snapshot of the head and then removes exactly the
// entries of that snapshot; a put, an overwrite or a tombstone written after the snapshot stays in the tree and in the
// WAL, so that a replay rebuilds the same head. The rows put the mutation after the snapshot in the order in which a
// writer without the lock of the column may run during a compaction; the counter rows have nothing after it.
func TestResetFoldedKeepsWhatCameAfterTheSnapshot(t *testing.T) {
	type entry struct {
		ts   int64
		val  int64
		str  string
		tomb bool
	}
	type step func(h *head) error
	put := func(ts, val int64, str string) step {
		return func(h *head) error { return h.put(ts, val, str) }
	}
	tomb := func(ts int64) step { return func(h *head) error { return h.tombstone(ts) } }
	for _, scheme := range []Scheme{SchemeDecimal, SchemeString} {
		for _, c := range []struct {
			name          string
			before, after []step
			want          []entry
		}{
			{"counter row: nothing after the snapshot", []step{put(1, 10, "a"), put(2, 20, "b")}, nil, nil},
			{"counter row: overwritten before the snapshot", []step{put(1, 10, "a"), put(1, 11, "b")}, nil, nil},
			{"correction after the snapshot", []step{put(1, 10, "a"), put(2, 20, "b")}, []step{put(3, 30, "c")},
				[]entry{{ts: 3, val: 30, str: "c"}}},
			{"tombstone after the snapshot", []step{put(1, 10, "a")}, []step{tomb(1)}, []entry{{ts: 1, tomb: true}}},
			{"overwrite after the snapshot", []step{put(1, 10, "a")}, []step{put(1, 11, "b")},
				[]entry{{ts: 1, val: 11, str: "b"}}},
			{"several after the snapshot", []step{put(5, 50, "e")}, []step{put(2, 20, "b"), tomb(5), put(2, 21, "c")},
				[]entry{{ts: 2, val: 21, str: "c"}, {ts: 5, tomb: true}}},
		} {
			t.Run(schemeName(scheme)+"/"+c.name, func(t *testing.T) {
				dir := t.TempDir()
				pool := ndb.NewFilePool(4)
				h, err := openHead(dir, pool, scheme)
				if err != nil {
					t.Fatal(err)
				}
				for _, s := range c.before {
					if err := s(h); err != nil {
						t.Fatal(err)
					}
				}
				mark := foldMark(h.snapshot(minInt64, maxInt64))
				for _, s := range c.after {
					if err := s(h); err != nil {
						t.Fatal(err)
					}
				}
				if err := h.resetFolded(mark); err != nil {
					t.Fatal(err)
				}
				check := func(h *head, when string) {
					t.Helper()
					got := h.snapshot(minInt64, maxInt64)
					if len(got) != len(c.want) {
						t.Fatalf("%s: %d entries %+v, want %+v", when, len(got), got, c.want)
					}
					tombs := 0
					for i, w := range c.want {
						g := got[i]
						if g.ts != w.ts || g.tombstone != w.tomb {
							t.Fatalf("%s: entry %d is %+v, want %+v", when, i, g, w)
						}
						if w.tomb {
							tombs++
							continue
						}
						if scheme == SchemeString && g.str != w.str || scheme != SchemeString && g.val != w.val {
							t.Fatalf("%s: entry %d is %+v, want %+v", when, i, g, w)
						}
					}
					if h.tombstoneCount() != tombs {
						t.Fatalf("%s: %d tombstones, want %d", when, h.tombstoneCount(), tombs)
					}
				}
				check(h, "after the fold")
				if len(c.want) == 0 && h.walSz != 0 {
					t.Fatalf("WAL of %d bytes after folding everything", h.walSz)
				}
				size := h.walSz
				if err := h.close(); err != nil {
					t.Fatal(err)
				}
				h, err = openHead(dir, pool, scheme)
				if err != nil {
					t.Fatal(err)
				}
				defer h.close()
				check(h, "after the replay")
				if h.walSz != size {
					t.Fatalf("replay read %d of %d bytes of the WAL", h.walSz, size)
				}
			})
		}
	}
}

func schemeName(s Scheme) string {
	if s == SchemeString {
		return "string"
	}
	return "decimal"
}

// TestCompactionKeepsTombstonesWrittenMeanwhile: tombstones of sealed points, which a writer puts without the lock of
// the column while compactions fold the head over and over, all survive the compactions, a last Flush and a reopen. The
// counter row deletes nothing, so every point stays. It checks the outcome under whatever interleaving the scheduler
// picks; against cd602526b, whose compaction cleared the whole head, deleted points came back.
func TestCompactionKeepsTombstonesWrittenMeanwhile(t *testing.T) {
	const n = 500
	for _, c := range []struct {
		name   string
		delete bool
		want   int
	}{
		{"counter row: nothing deleted", false, n},
		{"every point deleted while compacting", true, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			opts := Options{BlockPoints: 16, MaxHeadPoints: 4 * n}
			db, err := Open(dir, opts)
			if err != nil {
				t.Fatal(err)
			}
			col, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal})
			if err != nil {
				t.Fatal(err)
			}
			for ts := int64(1); ts <= n; ts++ {
				if err := col.PutI64(ts, ts); err != nil {
					t.Fatal(err)
				}
			}
			if err := col.Flush(); err != nil {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			done := make(chan struct{})
			var delErr, flushErr error
			wg.Go(func() {
				defer close(done)
				if !c.delete {
					return
				}
				for ts := int64(1); ts <= n; ts++ {
					if err := col.Delete(ts); err != nil {
						delErr = err
						return
					}
				}
			})
			wg.Go(func() {
				for {
					if err := col.Flush(); err != nil {
						flushErr = err
						return
					}
					select {
					case <-done:
						return
					default:
					}
				}
			})
			wg.Wait()
			if delErr != nil || flushErr != nil {
				t.Fatalf("delete: %v, flush: %v", delErr, flushErr)
			}
			check := func(col *Column, when string) {
				t.Helper()
				ts, _ := collectI64(t, col, 1, n)
				if len(ts) != c.want {
					t.Fatalf("%s: %d points, want %d", when, len(ts), c.want)
				}
			}
			check(col, "before the last flush")
			if err := col.Flush(); err != nil {
				t.Fatal(err)
			}
			check(col, "after the last flush")
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			db, err = Open(dir, opts)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			col, err = db.Column("b", "c", Schema{Scheme: SchemeDecimal})
			if err != nil {
				t.Fatal(err)
			}
			check(col, "after the reopen")
		})
	}
}
