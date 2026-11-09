// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// syncScheme writes and reads the points of the tests below in one scheme of a column. Point i has the time i*10.
type syncScheme struct {
	name   string
	schema Schema
	put    func(c *Column, i int) error
	want   func(i int) string
	read   func(t *testing.T, c *Column) ([]int64, []string)
}

func syncSchemes() []syncScheme {
	text := func(i int) string { return "v" + strconv.Itoa(i%3) }
	readStrings := func(t *testing.T, c *Column) ([]int64, []string) {
		t.Helper()
		return collectStrings(t, c, minInt64, maxInt64)
	}
	return []syncScheme{
		{
			name:   "decimal",
			schema: Schema{Scheme: SchemeDecimal},
			put:    func(c *Column, i int) error { return c.PutI64(int64(i)*10, int64(i)+1) },
			want:   func(i int) string { return strconv.Itoa(i + 1) },
			read: func(t *testing.T, c *Column) ([]int64, []string) {
				t.Helper()
				ts, vals := collectI64(t, c, minInt64, maxInt64)
				res := make([]string, len(vals))
				for i, v := range vals {
					res[i] = strconv.FormatInt(v, 10)
				}
				return ts, res
			},
		},
		{
			name:   "enum",
			schema: Schema{Scheme: SchemeEnum},
			put:    func(c *Column, i int) error { return c.PutString(int64(i)*10, text(i)) },
			want:   text,
			read:   readStrings,
		},
		{
			name:   "string",
			schema: Schema{Scheme: SchemeString},
			put:    func(c *Column, i int) error { return c.PutString(int64(i)*10, text(i)) },
			want:   text,
			read:   readStrings,
		},
	}
}

// checkSyncPoints expects exactly the points of the given indexes in ascending order.
func checkSyncPoints(t *testing.T, s syncScheme, c *Column, indexes []int) {
	t.Helper()
	ts, vals := s.read(t, c)
	if len(ts) != len(indexes) {
		t.Fatalf("%d points, want %d: %v", len(ts), len(indexes), ts)
	}
	for k, i := range indexes {
		if ts[k] != int64(i)*10 || vals[k] != s.want(i) {
			t.Fatalf("point %d is %d=%s, want %d=%s", k, ts[k], vals[k], int64(i)*10, s.want(i))
		}
	}
}

func upTo(n int) []int {
	res := make([]int, n)
	for i := range res {
		res[i] = i
	}
	return res
}

// crashCopy is the state of the disk after a crash of the process: a copy of the directory, taken without Flush and
// without Close, whose database has never written the memory of the running one.
func crashCopy(t *testing.T, dir string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "crashed")
	if err := os.CopyFS(dst, os.DirFS(dir)); err != nil {
		t.Fatal(err)
	}
	return dst
}

func reopenColumn(t *testing.T, dir string, opts Options) *Column {
	t.Helper()
	db, err := Open(dir, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	c, ok, err := db.LookupColumn("b", "c")
	if err != nil || !ok {
		t.Fatalf("reopen: ok=%v err=%v", ok, err)
	}
	return c
}

func syncChunkFiles(t *testing.T, dir string) (pending, final []string) {
	t.Helper()
	names, err := filepath.Glob(filepath.Join(dir, "b", "c", "*"+chunkExt))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range names {
		if strings.HasSuffix(filepath.Base(p), "_"+chunkExt) {
			pending = append(pending, p)
		} else {
			final = append(final, p)
		}
	}
	return pending, final
}

// TestSyncSurvivesCrash: what a Sync has written is there after a crash, what came after it is not. A block has 4 points,
// so 3 points are below a block and 10 points are two blocks and 2 buffered points.
func TestSyncSurvivesCrash(t *testing.T) {
	const appendedAfterCrash = 2
	tests := []struct {
		name      string
		points    int
		syncAfter int // the number of points after which Sync runs, 0 for no Sync
		want      int // points after the crash
	}{
		{name: "no sync: what is buffered is gone", points: 3, syncAfter: 0, want: 0},
		{name: "sync of a single point", points: 1, syncAfter: 1, want: 1},
		{name: "sync of less than a block", points: 3, syncAfter: 3, want: 3},
		{name: "points after the sync are gone", points: 5, syncAfter: 3, want: 3},
		{name: "no sync: full blocks stay, the rest is gone", points: 10, syncAfter: 0, want: 8},
		{name: "sync after full blocks keeps the rest", points: 10, syncAfter: 10, want: 10},
		{name: "sync at a block boundary has nothing to add", points: 8, syncAfter: 8, want: 8},
	}
	for _, s := range syncSchemes() {
		for _, tt := range tests {
			t.Run(s.name+"/"+tt.name, func(t *testing.T) {
				opts := Options{BlockPoints: 4}
				dir := t.TempDir()
				db, err := Open(dir, opts)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = db.Close() })
				c, err := db.Column("b", "c", s.schema)
				if err != nil {
					t.Fatal(err)
				}
				for i := range tt.points {
					if err := s.put(c, i); err != nil {
						t.Fatal(err)
					}
					if i+1 == tt.syncAfter {
						if err := c.Sync(); err != nil {
							t.Fatal(err)
						}
					}
				}
				checkSyncPoints(t, s, c, upTo(tt.points)) // the running database still has all of them

				crashed := crashCopy(t, dir)
				if pending, final := syncChunkFiles(t, crashed); len(pending) != 1 || len(final) != 0 {
					t.Fatalf("premise: the crash has to find an open chunk: pending %v, final %v", pending, final)
				}
				c2 := reopenColumn(t, crashed, opts)
				checkSyncPoints(t, s, c2, upTo(tt.want))

				// the column goes on after the recovery, and a second crash still has the synced points of both lives
				next := tt.points
				for i := range appendedAfterCrash {
					if err := s.put(c2, next+i); err != nil {
						t.Fatal(err)
					}
				}
				if err := c2.Sync(); err != nil {
					t.Fatal(err)
				}
				wantAll := append(upTo(tt.want), next, next+1)
				checkSyncPoints(t, s, c2, wantAll)
				checkSyncPoints(t, s, reopenColumn(t, crashCopy(t, crashed), opts), wantAll)
			})
		}
	}
}

// TestSyncTornTail: a crash in the middle of the write of a Sync costs this Sync only, not the blocks before it.
func TestSyncTornTail(t *testing.T) {
	tests := []struct {
		name string
		cut  int64 // bytes missing at the end of the open chunk
		want int
	}{
		{name: "nothing torn", cut: 0, want: 3},
		{name: "last block torn", cut: 3, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{BlockPoints: 4}
			dir := t.TempDir()
			db, err := Open(dir, opts)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			c, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal})
			if err != nil {
				t.Fatal(err)
			}
			s := syncSchemes()[0]
			for i := range 3 { // three blocks of one point
				if err := s.put(c, i); err != nil {
					t.Fatal(err)
				}
				if err := c.Sync(); err != nil {
					t.Fatal(err)
				}
			}
			crashed := crashCopy(t, dir)
			pending, _ := syncChunkFiles(t, crashed)
			if len(pending) != 1 {
				t.Fatalf("premise: open chunks: %v", pending)
			}
			fi, err := os.Stat(pending[0])
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Truncate(pending[0], fi.Size()-tt.cut); err != nil {
				t.Fatal(err)
			}
			checkSyncPoints(t, s, reopenColumn(t, crashed, opts), upTo(tt.want))
		})
	}
}

// TestSyncKeepsOneFile: however often Sync runs, the column has one open chunk file; Flush in the same place seals one
// file each time, which is why Sync exists.
func TestSyncKeepsOneFile(t *testing.T) {
	const rounds = 300
	tests := []struct {
		name        string
		write       func(*Column) error
		wantPending int
		wantFinal   int
	}{
		{name: "Sync", write: (*Column).Sync, wantPending: 1, wantFinal: 0},
		{name: "Flush", write: (*Column).Flush, wantPending: 0, wantFinal: rounds},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db, err := Open(dir, Options{})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			c, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal})
			if err != nil {
				t.Fatal(err)
			}
			s := syncSchemes()[0]
			for i := range rounds {
				if err := s.put(c, i); err != nil {
					t.Fatal(err)
				}
				if err := tt.write(c); err != nil {
					t.Fatal(err)
				}
			}
			pending, final := syncChunkFiles(t, dir)
			if len(pending) != tt.wantPending || len(final) != tt.wantFinal {
				t.Errorf("%d open and %d sealed files, want %d and %d", len(pending), len(final), tt.wantPending, tt.wantFinal)
			}
			checkSyncPoints(t, s, c, upTo(rounds))
		})
	}
}

// TestSyncEmptyBufferWritesNothing: with nothing buffered the open chunk does not grow; a new point does let it grow.
func TestSyncEmptyBufferWritesNothing(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	c, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Sync(); err != nil { // a column without any point
		t.Fatalf("sync of an empty column: %v", err)
	}
	if pending, final := syncChunkFiles(t, dir); len(pending)+len(final) != 0 {
		t.Fatalf("a sync without points made files: %v %v", pending, final)
	}

	size := func() int64 {
		t.Helper()
		pending, _ := syncChunkFiles(t, dir)
		if len(pending) != 1 {
			t.Fatalf("open chunks: %v", pending)
		}
		fi, err := os.Stat(pending[0])
		if err != nil {
			t.Fatal(err)
		}
		return fi.Size()
	}
	s := syncSchemes()[0]
	if err := s.put(c, 0); err != nil {
		t.Fatal(err)
	}
	if err := c.Sync(); err != nil {
		t.Fatal(err)
	}
	first := size()
	for range 3 {
		if err := c.Sync(); err != nil {
			t.Fatal(err)
		}
	}
	if got := size(); got != first {
		t.Errorf("syncs without a new point changed the chunk from %d to %d bytes", first, got)
	}
	if err := s.put(c, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Sync(); err != nil {
		t.Fatal(err)
	}
	if got := size(); got <= first {
		t.Errorf("sync of a new point left the chunk at %d bytes, it had %d", got, first)
	}
}

// TestSyncLetsTheSplitPolicyDecide: the split policy is asked only with an empty block buffer. After a Sync it is empty, so
// the next append can end the chunk; Sync itself never does.
func TestSyncLetsTheSplitPolicyDecide(t *testing.T) {
	tests := []struct {
		name        string
		sync        bool
		wantPending int
		wantFinal   int
	}{
		{name: "with Sync the policy ends the chunks", sync: true, wantPending: 1, wantFinal: 2},
		{name: "without Sync the buffer is never empty", sync: false, wantPending: 1, wantFinal: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db, err := Open(dir, Options{BlockPoints: 100, Split: SplitByCount(2)})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			c, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal})
			if err != nil {
				t.Fatal(err)
			}
			s := syncSchemes()[0]
			for i := range 6 {
				if err := s.put(c, i); err != nil {
					t.Fatal(err)
				}
				if tt.sync {
					if err := c.Sync(); err != nil {
						t.Fatal(err)
					}
				}
			}
			pending, final := syncChunkFiles(t, dir)
			if len(pending) != tt.wantPending || len(final) != tt.wantFinal {
				t.Errorf("%d open and %d sealed files, want %d and %d", len(pending), len(final), tt.wantPending, tt.wantFinal)
			}
			checkSyncPoints(t, s, c, upTo(6))
		})
	}
}

// TestSyncParallelWriters: every writer syncs and reads itself between its points, so that Sync runs against the writes
// and reads of the others without any assumption on the scheduler. Run with -race. Each point is there exactly once, the
// running column has one open chunk, and a crash after the last Sync keeps them all.
func TestSyncParallelWriters(t *testing.T) {
	const writers, perWriter = 8, 500
	opts := Options{BlockPoints: 16}
	dir := t.TempDir()
	db, err := Open(dir, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	c, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal})
	if err != nil {
		t.Fatal(err)
	}

	var next atomic.Int64
	var wg sync.WaitGroup
	errs := make([]error, writers)
	for w := range writers {
		wg.Go(func() {
			for i := range perWriter {
				ts := next.Add(1)
				if err := c.PutI64(ts, ts); err != nil {
					errs[w] = err
					return
				}
				if i%5 == 0 {
					if err := c.Sync(); err != nil {
						errs[w] = err
						return
					}
				}
				if i%50 == 0 {
					if err := c.ScanI64(minInt64, maxInt64, func(_, _ []int64) bool { return true }); err != nil {
						errs[w] = err
						return
					}
				}
			}
		})
	}
	wg.Wait()
	for w, err := range errs {
		if err != nil {
			t.Fatalf("writer %d: %v", w, err)
		}
	}
	if err := c.Sync(); err != nil {
		t.Fatal(err)
	}

	want := make([]int64, writers*perWriter)
	for i := range want {
		want[i] = int64(i) + 1
	}
	check := func(name string, c *Column) {
		t.Helper()
		ts, vals := collectI64(t, c, minInt64, maxInt64)
		if !slices.Equal(ts, vals) {
			t.Errorf("%s: the values differ from their times", name)
		}
		// a writer can lose the race for its later time, its point is then an out-of-order one of the head
		slices.Sort(ts)
		if !slices.Equal(ts, want) {
			t.Errorf("%s: %s", name, describeDifference(ts, want))
		}
	}
	check("running", c)
	if pending, final := syncChunkFiles(t, dir); len(pending) != 1 || len(final) != 0 {
		t.Errorf("%d open and %d sealed files, want one open file", len(pending), len(final))
	}
	check("after the crash", reopenColumn(t, crashCopy(t, dir), opts))
}

func describeDifference(got, want []int64) string {
	seen := make(map[int64]int, len(got))
	for _, ts := range got {
		seen[ts]++
	}
	var missing, twice int
	for _, ts := range want {
		switch seen[ts] {
		case 0:
			missing++
		case 1:
		default:
			twice++
		}
	}
	return fmt.Sprintf("%d points of %d: %d missing, %d more than once", len(got), len(want), missing, twice)
}
