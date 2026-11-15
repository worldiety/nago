// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// columnBytes is the size of all files of the column b/c.
func columnBytes(t testing.TB, dir string) int64 {
	t.Helper()
	return dirBytes(readSnapshot(t, filepath.Join(dir, "b", "c")))
}

// TestSyncAddsLittleToTheFiles: points after a Sync continue the open block. A sealed block makes the files of the column
// at most 64 B larger than the same points without a Sync between them, and each secured point of the open block makes
// them at most 48 B larger. The reference writes the same points without any Sync. Both end with Close. Close seals the
// open block as a block of its own, as today: its row shows that the comparison sees the frame of one more block.
func TestSyncAddsLittleToTheFiles(t *testing.T) {
	const bp = 16
	opts := tsdb.Options{BlockPoints: bp}
	for _, s := range surfaceSchemes() {
		write := func(t *testing.T, dir string, n int, syncAfter func(i int) bool, closeAfter int) {
			t.Helper()
			db := openDB(t, dir, opts)
			c := column(t, db, "b", "c", s.schema)
			for i := range n {
				must(t, s.put(c, int64(i)*10, value(s, i, 0)))
				if syncAfter(i) {
					must(t, c.Sync())
				}
				if i+1 == closeAfter {
					closeDB(t, db)
					db = openDB(t, dir, opts)
					c = lookup(t, db, "b", "c")
				}
			}
			closeDB(t, db) // seals what is left of a block; at a block boundary nothing is
		}
		never := func(int) bool { return false }
		every := func(k int) func(int) bool { return func(i int) bool { return (i+1)%k == 0 } }

		sealed := []struct {
			name       string
			syncAfter  func(int) bool
			closeAfter int
			atMost     int64 // more bytes than the reference; -1 for no upper bound
			atLeast    int64
		}{
			{name: "a Sync after each point", syncAfter: every(1), atMost: 3 * 64},
			{name: "a Sync after every 5 points", syncAfter: every(5), atMost: 3 * 64},
			{name: "a Sync after every 7 points", syncAfter: every(7), atMost: 3 * 64},
			{name: "counter row: Close in the middle of a block seals it", syncAfter: never, closeAfter: 8, atMost: -1, atLeast: 1},
		}
		for _, tt := range sealed {
			t.Run(s.name+"/3 sealed blocks/"+tt.name, func(t *testing.T) {
				ref, dir := t.TempDir(), t.TempDir()
				write(t, ref, 3*bp, never, 0)
				write(t, dir, 3*bp, tt.syncAfter, tt.closeAfter)
				extra := columnBytes(t, dir) - columnBytes(t, ref)
				t.Logf("%d B more than without Sync", extra)
				if tt.atMost >= 0 && extra > tt.atMost {
					t.Errorf("3 sealed blocks take %d B more than without Sync, at most %d", extra, tt.atMost)
				}
				if extra < tt.atLeast {
					t.Errorf("%d B more than without Sync, at least %d", extra, tt.atLeast)
				}
			})
		}

		open := []struct {
			name      string
			before    int // points before the open block
			closed    bool
			points    int // secured points of the open block, measured; all of them stay below a block
			syncAfter func(int) bool
		}{
			// the first point makes the files of the chunk, the bound is for the points after it
			{name: "a new column, a Sync after each point", before: 1, points: 10, syncAfter: every(1)},
			{name: "after a sealed block, a Sync after each point", before: bp, points: 10, syncAfter: every(1)},
			{name: "after a sealed block, a Sync after every 3 points", before: bp, points: 12, syncAfter: every(3)},
			{name: "after Close, a Sync after each point", before: bp + 2, closed: true, points: 10, syncAfter: every(1)},
		}
		for _, tt := range open {
			t.Run(s.name+"/the open block/"+tt.name, func(t *testing.T) {
				dir := t.TempDir()
				db := openDB(t, dir, opts)
				c := column(t, db, "b", "c", s.schema)
				for i := range tt.before {
					must(t, s.put(c, int64(i)*10, value(s, i, 0)))
				}
				must(t, c.Sync())
				if tt.closed {
					closeDB(t, db)
					db = openDB(t, dir, opts)
					c = lookup(t, db, "b", "c")
				}
				before := columnBytes(t, dir)
				for i := range tt.points {
					must(t, s.put(c, int64(tt.before+i)*10, value(s, tt.before+i, 0)))
					if tt.syncAfter(i) || i == tt.points-1 {
						must(t, c.Sync())
					}
				}
				if u := c.Stats().Unsynced; u != 0 {
					t.Fatalf("premise: %d points in memory after the last Sync", u)
				}
				extra := columnBytes(t, dir) - before
				t.Logf("%d secured points of the open block take %d B", tt.points, extra)
				if extra > int64(48*tt.points) {
					t.Errorf("%d secured points of the open block take %d B, at most %d", tt.points, extra, 48*tt.points)
				}
			})
		}
	}
}

// splitCall is what a split policy saw: the index of the point about to be written (its time / 10), the statistics it
// got, and the size of the open chunk file at that moment.
type splitCall struct {
	index      int64
	points     int64
	size, file int64
}

// recordingSplit records each call with points > 0 and ends the chunk when end says so.
func recordingSplit(t *testing.T, dir string, end func(tsdb.ChunkStats) bool) (tsdb.SplitFunc, func() []splitCall) {
	var mu sync.Mutex
	var calls []splitCall
	return func(st tsdb.ChunkStats) bool {
			if st.Points == 0 {
				return false
			}
			open, _ := filepath.Glob(filepath.Join(dir, "b", "c", "*_.tsb"))
			var file int64 = -1
			if len(open) == 1 {
				if fi, err := os.Stat(open[0]); err == nil {
					file = fi.Size()
				}
			}
			mu.Lock()
			calls = append(calls, splitCall{index: st.NextMillis / 10, points: st.Points, size: st.SizeBytes, file: file})
			mu.Unlock()
			return end != nil && end(st)
		}, func() []splitCall {
			mu.Lock()
			defer mu.Unlock()
			return slices.Clone(calls)
		}
}

// TestSplitPolicyIsAskedAfterSyncAndAtBlocks: the split policy is asked at the first point after a Sync and after the
// open, as today, and otherwise at block boundaries; a block continued over a Sync ends after BlockPoints points. It
// gets every secured point of the chunk in Points and the size of its file in SizeBytes. A block has 4 points, so
// without a Sync the policy is asked at the points 4 and 8; a Sync after point 1 adds the ask at 2 and keeps 4 and 8.
// Close and a compaction of the open chunk seal the open block as a block of its own, so the next block begins after
// them.
func TestSplitPolicyIsAskedAfterSyncAndAtBlocks(t *testing.T) {
	const points = 10
	var rewriting atomic.Bool
	for _, tt := range []struct {
		name string
		// between runs after the point of each index
		between func(t *testing.T, i int, c *tsdb.Column, reopen func() *tsdb.Column) *tsdb.Column
		opts    tsdb.Options
		want    []int64 // the indexes of the points at which the policy is asked
	}{
		{name: "no Sync", want: []int64{4, 8}},
		{name: "a Sync after point 1", want: []int64{2, 4, 8}, between: syncAt(1)},
		{name: "a Sync after the points 1 and 2", want: []int64{2, 3, 4, 8}, between: syncAt(1, 2)},
		{name: "a Sync after point 5", want: []int64{4, 6, 8}, between: syncAt(5)},
		{name: "a Sync at the block boundary", want: []int64{4, 8}, between: syncAt(3)},
		{name: "Close after point 5", want: []int64{4, 6}, between: func(t *testing.T, i int, c *tsdb.Column, reopen func() *tsdb.Column) *tsdb.Column {
			if i == 5 {
				return reopen()
			}
			return c
		}},
		{name: "a compaction of the open chunk after point 5", opts: tsdb.Options{MaxHeadPoints: 1}, want: []int64{4, 6},
			between: func(t *testing.T, i int, c *tsdb.Column, _ func() *tsdb.Column) *tsdb.Column {
				if i == 5 {
					// a correction in the open chunk; the full head compacts at once. What the rewrite asks the
					// policy is not counted: only the appends are.
					rewriting.Store(true)
					must(t, c.PutI64(20, -1))
					rewriting.Store(false)
				}
				return c
			}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			split, calls := recordingSplit(t, dir, nil)
			opts := tt.opts
			appendSplit := split
			split = func(st tsdb.ChunkStats) bool {
				if rewriting.Load() {
					return false
				}
				return appendSplit(st)
			}
			opts.BlockPoints, opts.Split = 4, split
			db := openDB(t, dir, opts)
			c := column(t, db, "b", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
			reopen := func() *tsdb.Column {
				closeDB(t, db)
				db = openDB(t, dir, opts)
				return lookup(t, db, "b", "c")
			}
			for i := range points {
				must(t, c.PutI64(int64(i)*10, int64(i)))
				if tt.between != nil {
					c = tt.between(t, i, c, reopen)
				}
			}
			got := calls()
			var at []int64
			for _, call := range got {
				at = append(at, call.index)
				if call.points != call.index {
					t.Errorf("at point %d the policy got Points %d, want the %d secured points before it", call.index, call.points, call.index)
				}
				if call.size != call.file {
					t.Errorf("at point %d the policy got SizeBytes %d, the open chunk file has %d", call.index, call.size, call.file)
				}
			}
			if !slices.Equal(at, tt.want) {
				t.Errorf("the policy was asked at the points %v, want %v", at, tt.want)
			}
		})
	}
}

func syncAt(indexes ...int) func(t *testing.T, i int, c *tsdb.Column, _ func() *tsdb.Column) *tsdb.Column {
	return func(t *testing.T, i int, c *tsdb.Column, _ func() *tsdb.Column) *tsdb.Column {
		if slices.Contains(indexes, i) {
			must(t, c.Sync())
		}
		return c
	}
}

// TestSplitPolicyEndsTheChunkAtASync: when the policy ends the chunk at the first point after a Sync, the open block is
// sealed there: the finalized chunk holds exactly the points before, the point begins the new open chunk, and both
// survive a crash after the next Sync. The counter row without Sync shows that the policy is not asked there.
func TestSplitPolicyEndsTheChunkAtASync(t *testing.T) {
	for _, tt := range []struct {
		name      string
		sync      bool
		wantFinal []string
	}{
		{name: "Sync after point 2", sync: true, wantFinal: []string{"0_20.tsb"}},
		{name: "counter row: no Sync", sync: false, wantFinal: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			split, _ := recordingSplit(t, dir, func(st tsdb.ChunkStats) bool { return st.Points >= 3 })
			opts := tsdb.Options{BlockPoints: 64, Split: split}
			db := openDB(t, dir, opts)
			c := column(t, db, "b", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
			for i := range 6 {
				must(t, c.PutI64(int64(i)*10, int64(i)))
				if tt.sync && i == 2 {
					must(t, c.Sync())
				}
			}
			must(t, c.Sync())
			var final []string
			entries, err := os.ReadDir(filepath.Join(dir, "b", "c"))
			must(t, err)
			for _, e := range entries {
				if chunkName.MatchString(e.Name()) {
					final = append(final, e.Name())
				}
			}
			if !slices.Equal(final, tt.wantFinal) {
				t.Errorf("finalized chunks %v, want %v", final, tt.wantFinal)
			}
			want := map[int64]string{0: "0", 10: "1", 20: "2", 30: "3", 40: "4", 50: "5"}
			for name, got := range map[string]map[int64]string{
				"running":         readMap(t, decimalScheme(), c),
				"after the crash": readMap(t, decimalScheme(), lookup(t, openDB(t, crashCopy(t, dir), opts), "b", "c")),
			} {
				if !mapsEqual(got, want) {
					t.Errorf("%s: %s, want %s", name, describeMap(got), describeMap(want))
				}
			}
		})
	}
}

// TestUnsyncedCountsThePointsOnlyInMemory: ColumnStats.Unsynced is the number of points of a column that are only in
// memory. A Sync and a sealed block put them on disk, and a point after a Sync continues the open block, so the fourth
// point of a block of 4 seals it although a Sync came before. Corrections and deletions are in the head's log on disk.
func TestUnsyncedCountsThePointsOnlyInMemory(t *testing.T) {
	for _, s := range surfaceSchemes() {
		t.Run(s.name, func(t *testing.T) {
			dir := t.TempDir()
			opts := quietOptions(4)
			db := openDB(t, dir, opts)
			c := column(t, db, "b", "c", s.schema)
			put := func(i int) func() error { return func() error { return s.put(c, int64(i)*10, value(s, i, 0)) } }
			puts := func(from, to int) func() error {
				return func() error {
					for i := from; i < to; i++ {
						if err := put(i)(); err != nil {
							return err
						}
					}
					return nil
				}
			}
			steps := []struct {
				name string
				do   func() error
				want int64
			}{
				{name: "a new column", do: func() error { return nil }, want: 0},
				{name: "3 points", do: puts(0, 3), want: 3},
				{name: "Sync", do: c.Sync, want: 0},
				{name: "the fourth point of the block seals it", do: put(3), want: 0},
				{name: "a point of the next block", do: put(4), want: 1},
				{name: "a correction of a point on disk", do: func() error { return s.put(c, 10, value(s, 1, 1)) }, want: 1},
				{name: "a deletion", do: func() error { return c.Delete(20) }, want: 1},
				{name: "a Sync of one point", do: c.Sync, want: 0},
				{name: "two more points", do: puts(5, 7), want: 2},
				{name: "the block is full", do: put(7), want: 0},
				{name: "a full block without Sync", do: puts(8, 12), want: 0},
				{name: "2 points", do: puts(12, 14), want: 2},
			}
			for _, st := range steps {
				must(t, st.do())
				if got := c.Stats().Unsynced; got != st.want {
					t.Errorf("%s: Unsynced %d, want %d", st.name, got, st.want)
				}
			}
			crashed := lookup(t, openDB(t, crashCopy(t, dir), opts), "b", "c")
			if got := crashed.Stats().Unsynced; got != 0 {
				t.Errorf("after a crash: Unsynced %d, want 0", got)
			}
			closeDB(t, db)
			if got := lookup(t, openDB(t, dir, opts), "b", "c").Stats().Unsynced; got != 0 {
				t.Errorf("after Close: Unsynced %d, want 0", got)
			}
		})
	}
}

// readOnceModel holds the points a column must have: exactly the secured ones, and of the unsecured ones those a
// running column still has or a crash may have lost.
type readOnceModel struct {
	secured, unsecured map[int64]string
}

func (m readOnceModel) put(ts int64, v string, secured bool) {
	if secured {
		m.secured[ts] = v
		delete(m.unsecured, ts)
		return
	}
	m.unsecured[ts] = v
}

func (m readOnceModel) sync() {
	for ts, v := range m.unsecured {
		m.secured[ts] = v
	}
	clear(m.unsecured)
}

// TestEachPointIsReadOnce: every secured point that is not deleted is read exactly once and no point more than once,
// while the column runs, after Close and an open, and after a crash, by each read of the column and over ranges that
// begin and end inside the open block, between its secured and its unsecured points and in a gap filled by a
// correction. A block has 8 points; the open block is continued by Syncs, corrected through the head, a gap in it and one
// between its points in memory are filled, a point is deleted. The running column and the one after Close read exactly
// the model; after a crash the unsecured points may be missing.
func TestEachPointIsReadOnce(t *testing.T) {
	const bp = 8
	opts := quietOptions(bp)
	for _, s := range surfaceSchemes() {
		t.Run(s.name, func(t *testing.T) {
			dir := t.TempDir()
			db := openDB(t, dir, opts)
			c := column(t, db, "b", "c", s.schema)
			m := readOnceModel{secured: map[int64]string{}, unsecured: map[int64]string{}}
			put := func(i int, gen int, ts int64, secured bool) {
				v := value(s, i, gen)
				must(t, s.put(c, ts, v))
				m.put(ts, v, secured)
			}
			for i := range bp + 3 { // a sealed block and 3 points of the open block
				put(i, 0, int64(i)*10, i < bp)
			}
			must(t, c.Sync())
			m.sync()
			put(11, 0, 110, false)
			put(12, 0, 120, false)
			must(t, c.Sync())
			m.sync()
			put(1, 1, 90, true)    // a correction of a secured point of the open block
			put(2, 2, 95, true)    // a new point in a gap of the secured part of the open block
			put(3, 3, 30, true)    // a correction in the sealed block
			put(13, 0, 130, false) // points in memory
			put(14, 0, 140, false)
			put(4, 4, 135, true) // a new point between points in memory
			must(t, c.Delete(100))
			delete(m.secured, 100)

			check := func(t *testing.T, life string, c *tsdb.Column, exact bool) {
				ranges := [][2]int64{
					{minTS, maxTS}, {0, 140}, {80, 140}, {85, 125}, {95, 95}, {90, 100}, {100, 100}, {111, 139},
					{125, 135}, {130, 140}, {135, 135}, {-100, -1}, {141, 1 << 40}, {25, 35},
				}
				for _, r := range ranges {
					for name, got := range readAll(t, s, c, r[0], r[1]) {
						where := fmt.Sprintf("%s, %s over [%d, %d]", life, name, r[0], r[1])
						checkReadOnce(t, where, got, m, r, exact)
					}
				}
				n, err := c.Count()
				must(t, err)
				if exact && n != int64(len(m.secured)+len(m.unsecured)) {
					t.Errorf("%s: Count %d, want %d", life, n, len(m.secured)+len(m.unsecured))
				}
				if !exact && (n < int64(len(m.secured)) || n > int64(len(m.secured)+len(m.unsecured))) {
					t.Errorf("%s: Count %d, want %d to %d", life, n, len(m.secured), len(m.secured)+len(m.unsecured))
				}
			}
			check(t, "running", c, true)
			crashed := crashCopy(t, dir)
			check(t, "after a crash", lookup(t, openDB(t, crashed, opts), "b", "c"), false)
			closeDB(t, db)
			m.sync()
			check(t, "after Close", lookup(t, openDB(t, dir, opts), "b", "c"), true)
		})
	}
}

// readAll reads the range by every read of the scheme, as time and value strings.
func readAll(t *testing.T, s surfaceScheme, c *tsdb.Column, min, max int64) map[string][][2]string {
	t.Helper()
	res := map[string][][2]string{}
	add := func(name string, ts int64, v string) {
		res[name] = append(res[name], [2]string{strconv.FormatInt(ts, 10), v})
	}
	if s.schema.Scheme == tsdb.SchemeDecimal {
		must(t, c.ScanI64(min, max, func(ts, vs []int64) bool {
			for i := range ts {
				add("ScanI64", ts[i], strconv.FormatInt(vs[i], 10))
			}
			return true
		}))
		must(t, c.ScanF64(min, max, func(ts []int64, vs []float64) bool {
			for i := range ts {
				add("ScanF64", ts[i], strconv.FormatInt(int64(vs[i]), 10))
			}
			return true
		}))
		for p := range c.IterI64(min, max) {
			add("IterI64", int64(p.X), strconv.FormatInt(int64(p.Y), 10))
		}
		for p := range c.IterF64(min, max) {
			add("IterF64", int64(p.X), strconv.FormatInt(int64(p.Y), 10))
		}
		return res
	}
	must(t, c.ScanString(min, max, func(ts []int64, vs []string) bool {
		for i := range ts {
			add("ScanString", ts[i], vs[i])
		}
		return true
	}))
	for p := range c.IterString(min, max) {
		add("IterString", int64(p.X), p.V)
	}
	return res
}

func checkReadOnce(t *testing.T, where string, got [][2]string, m readOnceModel, r [2]int64, exact bool) {
	t.Helper()
	seen := map[int64]string{}
	last := int64(minTS)
	for i, p := range got {
		ts, _ := strconv.ParseInt(p[0], 10, 64)
		if _, ok := seen[ts]; ok {
			t.Errorf("%s: the time %d is read twice", where, ts)
			return
		}
		if i > 0 && ts < last {
			t.Errorf("%s: the time %d after %d", where, ts, last)
			return
		}
		last = ts
		seen[ts] = p[1]
		if ts < r[0] || ts > r[1] {
			t.Errorf("%s: the time %d is outside the range", where, ts)
		}
	}
	for ts, v := range m.secured {
		if ts < r[0] || ts > r[1] {
			continue
		}
		if got, ok := seen[ts]; !ok || got != v {
			t.Errorf("%s: the secured time %d reads %q (present %v), want %q", where, ts, got, ok, v)
		}
	}
	for ts, got := range seen {
		if v, ok := m.secured[ts]; ok {
			_ = v
			continue
		}
		if v, ok := m.unsecured[ts]; !ok || v != got {
			t.Errorf("%s: the time %d reads %q, which was never written there or is deleted", where, ts, got)
		}
	}
	if exact {
		for ts := range m.unsecured {
			if _, ok := seen[ts]; !ok && ts >= r[0] && ts <= r[1] {
				t.Errorf("%s: the running column misses the point %d in memory", where, ts)
			}
		}
	}
}

// TestColumnKeepsItsFiles: after each Sync, Close, open and finished compaction a column has no more files than today:
// schema.json, head.wal, for enum the dictionary and its one open chunk, as long as the split policy ends no chunk. A
// compaction that rewrites the open chunk keeps it open. Flush finalizes the open chunk, so the next point begins a
// second file: its rows show that the count sees chunk files.
func TestColumnKeepsItsFiles(t *testing.T) {
	for _, s := range surfaceSchemes() {
		t.Run(s.name, func(t *testing.T) {
			enum := s.schema.Scheme == tsdb.SchemeEnum
			dir := t.TempDir()
			opts := tsdb.Options{BlockPoints: 8, MaxHeadPoints: 3}
			db := openDB(t, dir, opts)
			c := column(t, db, "b", "c", s.schema)
			colDir := filepath.Join(dir, "b", "c")
			files := func(step string, wantFinal, wantOpen int) {
				t.Helper()
				final, open := columnFiles(t, colDir, enum)
				if final != wantFinal || open != wantOpen {
					t.Errorf("after %s: %d finalized and %d open chunks, want %d and %d", step, final, open, wantFinal, wantOpen)
				}
			}
			n := 0
			put := func() { must(t, s.put(c, int64(n)*10, value(s, n, 0))); n++ }
			for range 30 {
				put()
				must(t, c.Sync())
				files(fmt.Sprintf("Sync of point %d", n-1), 0, 1)
			}
			closeDB(t, db)
			files("Close", 0, 1)
			db = openDB(t, dir, opts)
			c = lookup(t, db, "b", "c")
			files("the open", 0, 1)
			for i := range 3 { // the third correction fills the head, whose compaction rewrites the open chunk
				must(t, s.put(c, int64(25-i)*10, value(s, i, 1)))
			}
			files("a compaction of the open chunk", 0, 1)
			put()
			must(t, c.Sync())
			files("a Sync after the compaction", 0, 1)
			must(t, s.put(c, 30, value(s, 3, 1)))
			closeDB(t, db)
			files("Close with a correction", 0, 1)
			db = openDB(t, dir, opts)
			c = lookup(t, db, "b", "c")
			must(t, c.Flush())
			files("counter row: Flush", 1, 0)
			put()
			must(t, c.Sync())
			files("counter row: a Sync after Flush", 1, 1)
		})
	}
}
