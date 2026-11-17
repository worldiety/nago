// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The memory tests use the default of 4096 BlockPoints, so that a buffer of the size of a block can break the bounds, and
// few columns. They measure the live heap after a full collection, before the columns are made and after, and divide
// by the number of columns. What a column holds includes its handles in the FilePool.

const kib = 1024

// columns is the number of columns of a memory test: n, or with -short a quarter of it. The bounds are per column and
// break by tens of KiB against d9eaf8db2, which a quarter of the columns still shows.
func columns(n int) int {
	if testing.Short() {
		return n / 4
	}
	return n
}

// putSeries writes n points of oscillating values from the index from on, one each 10 ms.
func putSeries(t testing.TB, c *tsdb.Column, from, n int) {
	t.Helper()
	for i := from; i < from+n; i++ {
		must(t, c.PutI64(int64(i)*10, int64(1000+i*37%101)))
	}
}

func seriesName(i int) string { return fmt.Sprintf("s%04d", i) }

// perColumn measures the heap of n columns, which fill makes and keeps, against the heap before.
func perColumn(t testing.TB, n int, fill func() []*tsdb.Column) int64 {
	t.Helper()
	before := liveHeap()
	cols := fill()
	after := liveHeap()
	if len(cols) != n {
		t.Fatalf("premise: %d columns, want %d", len(cols), n)
	}
	runtime.KeepAlive(cols)
	return (after - before) / int64(n)
}

func unsyncedOf(cols []*tsdb.Column) (lo, hi int64) {
	lo, hi = -1, -1
	for _, c := range cols {
		u := c.Stats().Unsynced
		if lo < 0 || u < lo {
			lo = u
		}
		if u > hi {
			hi = u
		}
	}
	return lo, hi
}

// TestColumnMemoryAfterSync: after Sync, after the sealing of a block and after the open a column holds no memory that
// grows with the size of a block: 200 columns with 4095 points each and a Sync take at most 8 KiB each. Unsecured points
// take at most 24 B each on top: the row without Sync, which holds 4095 points in memory, shows that the measurement sees
// them. With -short a quarter of the columns, and the rows that d9eaf8db2 already passed run without -short only.
func TestColumnMemoryAfterSync(t *testing.T) {
	cols := columns(200)
	opts := tsdb.Options{}
	for _, tt := range []struct {
		name         string
		points       int
		sync         bool
		reopen       string // "", "Close" or "crash": measure the columns opened after it
		wantUnsynced int64
		long         bool // only without -short
	}{
		{name: "4095 points and Sync", points: 4095, sync: true, wantUnsynced: 0},
		{name: "4096 points: a sealed block", points: 4096, wantUnsynced: 0},
		{name: "4097 points: a sealed block and one point", points: 4097, wantUnsynced: 1},
		{name: "4095 points without Sync", points: 4095, wantUnsynced: 4095},
		{name: "2000 points without Sync", points: 2000, wantUnsynced: 2000, long: true},
		{name: "4095 points, Close and the open of each column", points: 4095, reopen: "Close", wantUnsynced: 0},
		{name: "4095 points, Sync, a crash and the open of each column", points: 4095, sync: true, reopen: "crash", wantUnsynced: 0, long: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.long && testing.Short() {
				t.Skip("-short: d9eaf8db2 passes this row already")
			}
			dir := t.TempDir()
			db := openDB(t, dir, opts)
			write := func() []*tsdb.Column {
				res := make([]*tsdb.Column, cols)
				for i := range res {
					res[i] = column(t, db, "b", seriesName(i), tsdb.Schema{Scheme: tsdb.SchemeDecimal})
					putSeries(t, res[i], 0, tt.points)
					if tt.sync {
						must(t, res[i].Sync())
					}
				}
				return res
			}
			var per int64
			var got []*tsdb.Column
			if tt.reopen == "" {
				per = perColumn(t, cols, func() []*tsdb.Column { got = write(); return got })
			} else {
				write()
				from := dir
				if tt.reopen == "crash" {
					from = crashCopy(t, dir)
				} else {
					closeDB(t, db)
				}
				db2 := openDB(t, from, opts)
				per = perColumn(t, cols, func() []*tsdb.Column {
					got = make([]*tsdb.Column, cols)
					for i := range got {
						got[i] = lookup(t, db2, "b", seriesName(i))
					}
					return got
				})
			}
			lo, hi := unsyncedOf(got)
			if lo != tt.wantUnsynced || hi != tt.wantUnsynced {
				t.Errorf("Unsynced %d to %d, want %d", lo, hi, tt.wantUnsynced)
			}
			limit := 8*kib + 24*tt.wantUnsynced
			t.Logf("%d B per column with %d points in memory, at most %d", per, tt.wantUnsynced, limit)
			if per > limit {
				t.Errorf("%d B per column with %d points in memory, at most %d", per, tt.wantUnsynced, limit)
			}
		})
	}
}

// TestRestingColumnMemory: the head, the dictionary and the list of chunks stay in memory. A column without buffered
// points and with an empty head takes at most 8 KiB, plus 256 B for each chunk over 8 and 64 B for each entry of its
// dictionary plus its length. After a query it holds at most 4 KiB more for reading. A chunk ends after 10 points by the
// policy, which is asked after each Sync. With -short a quarter of the columns.
func TestRestingColumnMemory(t *testing.T) {
	for _, tt := range []struct {
		name    string
		cols    int
		schema  tsdb.Schema
		fill    func(t testing.TB, c *tsdb.Column)
		bound   int64
		opts    tsdb.Options
		chunks  int // per column, the open one included
		entries int // of the dictionary
	}{
		{
			name: "20 chunks", cols: columns(40), schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}, chunks: 20,
			opts:  tsdb.Options{Split: tsdb.SplitByCount(10)},
			bound: 8*kib + (20-8)*256,
			fill: func(t testing.TB, c *tsdb.Column) {
				for k := range 20 {
					putSeries(t, c, k*10, 10)
					must(t, c.Sync())
				}
			},
		},
		{
			name: "a dictionary of 100 names of 16 bytes", cols: columns(100), schema: tsdb.Schema{Scheme: tsdb.SchemeEnum}, chunks: 1, entries: 100,
			bound: 8*kib + 100*(64+16),
			fill: func(t testing.TB, c *tsdb.Column) {
				for i := range 100 {
					must(t, c.PutString(int64(i)*10, fmt.Sprintf("state-%010d", i)))
				}
				must(t, c.Sync())
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db := openDB(t, dir, tt.opts)
			var cols []*tsdb.Column
			per := perColumn(t, tt.cols, func() []*tsdb.Column {
				for i := range tt.cols {
					c := column(t, db, "b", seriesName(i), tt.schema)
					tt.fill(t, c)
					cols = append(cols, c)
				}
				return cols
			})
			if st := cols[0].Stats(); st.Chunks != tt.chunks || st.Unsynced != 0 {
				t.Fatalf("premise: %d chunks and %d points in memory, want %d and 0", st.Chunks, st.Unsynced, tt.chunks)
			}
			t.Logf("running: %d B per column, at most %d", per, tt.bound)
			if per > tt.bound {
				t.Errorf("running: %d B per column, at most %d", per, tt.bound)
			}

			closeDB(t, db)
			db2 := openDB(t, dir, tt.opts)
			per = perColumn(t, tt.cols, func() []*tsdb.Column {
				cols = cols[:0]
				for i := range tt.cols {
					c := lookup(t, db2, "b", seriesName(i))
					if n, err := c.Count(); err != nil || n == 0 {
						t.Fatalf("premise: the query found %d points: %v", n, err)
					}
					cols = append(cols, c)
				}
				return cols
			})
			t.Logf("after the open and a query: %d B per column, at most %d", per, tt.bound+4*kib)
			if per > tt.bound+4*kib {
				t.Errorf("after the open and a query: %d B per column, at most %d", per, tt.bound+4*kib)
			}
		})
	}
}

// TestColumnMemoryForReading: after a query a column holds at most 4 KiB for reading. 100 columns of three sealed blocks
// and an open block first answer a query of a range before all points, which needs no block, and then queries of all
// points and of a few; the heap between them grows by at most 4 KiB per column. The counter row holds the read points
// itself and shows that the measurement sees memory kept after a query. With -short a quarter of the columns.
func TestColumnMemoryForReading(t *testing.T) {
	cols := columns(100)
	for _, tt := range []struct {
		name   string
		reopen bool
		keep   bool
		schema tsdb.Schema
	}{
		{name: "decimal, running", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}},
		{name: "decimal, after Close and the open", reopen: true, schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}},
		{name: "enum, after Close and the open", reopen: true, schema: tsdb.Schema{Scheme: tsdb.SchemeEnum}},
		{name: "counter row: the test keeps what it read", keep: true, schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db := openDB(t, dir, tsdb.Options{})
			const points = 3*4096 + 100
			var cols0 []*tsdb.Column
			for i := range cols {
				c := column(t, db, "b", seriesName(i), tt.schema)
				for p := range points {
					if tt.schema.Scheme == tsdb.SchemeEnum {
						must(t, c.PutString(int64(p)*10, fmt.Sprint("state", p%5)))
					} else {
						must(t, c.PutI64(int64(p)*10, int64(p*37%1001)))
					}
				}
				must(t, c.Sync())
				cols0 = append(cols0, c)
			}
			if tt.reopen {
				closeDB(t, db)
				db = openDB(t, dir, tsdb.Options{})
				cols0 = cols0[:0]
				for i := range cols {
					cols0 = append(cols0, lookup(t, db, "b", seriesName(i)))
				}
			}
			count := func(c *tsdb.Column, min, max int64, keep *[]int64) int {
				n := 0
				if tt.schema.Scheme == tsdb.SchemeEnum {
					must(t, c.ScanString(min, max, func(ts []int64, _ []string) bool { n += len(ts); return true }))
					return n
				}
				must(t, c.ScanI64(min, max, func(ts, vs []int64) bool {
					n += len(ts)
					if keep != nil {
						*keep = append(*keep, vs...)
					}
					return true
				}))
				return n
			}
			for _, c := range cols0 {
				if n := count(c, -1000, -1, nil); n != 0 {
					t.Fatalf("premise: %d points before all points", n)
				}
			}
			kept := make([][]int64, cols)
			per := perColumn(t, cols, func() []*tsdb.Column {
				for i, c := range cols0 {
					var keep *[]int64
					if tt.keep {
						keep = &kept[i]
					}
					if n := count(c, minTS, maxTS, keep); n != points {
						t.Fatalf("premise: %d points, want %d", n, points)
					}
					if n := count(c, 20_000, 20_990, nil); n != 100 {
						t.Fatalf("premise: %d points in the few, want 100", n)
					}
				}
				return cols0
			})
			runtime.KeepAlive(kept)
			if tt.keep {
				t.Logf("counter row: %d B per column", per)
				if per < points*8 {
					t.Errorf("counter row: %d B per column, the test kept %d B itself", per, points*8)
				}
				return
			}
			t.Logf("%d B per column for reading, at most %d", per, 4*kib)
			if per > 4*kib {
				t.Errorf("%d B per column for reading, at most %d", per, 4*kib)
			}
		})
	}
}

// TestReleaseIdle: DB.ReleaseIdle secures every column, into which nothing was written for IdleAfter by the clock of
// Options.Now, as Sync does, and frees its memory; IdleAfter 0 is off. A column written again within IdleAfter stays as
// it is, also next to columns that are released; reading does not keep a column awake, and the released column keeps at
// most 4 KiB from that query. A held *Column stays valid and is the one DB.Column returns. 100 columns of 4000 points
// each are written at the time t0, with -short a quarter of them; the clock is the test's own, IdleAfter is an hour.
func TestReleaseIdle(t *testing.T) {
	const points = 4000
	cols := columns(100)
	const idle = time.Hour
	for _, tt := range []struct {
		name      string
		idleAfter time.Duration
		// between runs after the write of all columns at t0, with the clock at t0 + 30 min
		between func(t *testing.T, i int, c *tsdb.Column)
		wait    time.Duration // from t0 to ReleaseIdle
		want    func(i int) int64
		memory  int64 // the released columns take at most this each; 0 for no bound
	}{
		{name: "resting for IdleAfter and 1 ms", idleAfter: idle, wait: idle + time.Millisecond,
			want: func(int) int64 { return 0 }, memory: 8 * kib},
		{name: "counter row: resting for 1 ms less than IdleAfter", idleAfter: idle, wait: idle - time.Millisecond,
			want: func(int) int64 { return points }},
		{name: "counter row: IdleAfter 0 is off", idleAfter: 0, wait: 24 * time.Hour,
			want: func(int) int64 { return points }},
		{name: "only read since the write", idleAfter: idle, wait: idle + time.Millisecond,
			between: func(t *testing.T, _ int, c *tsdb.Column) {
				if n, err := c.Count(); err != nil || n != points {
					t.Fatalf("count %d %v", n, err)
				}
			},
			want: func(int) int64 { return 0 }, memory: 8*kib + 4*kib}, // and what a query keeps for reading
		{name: "counter row: written again since", idleAfter: idle, wait: idle + time.Millisecond,
			between: func(t *testing.T, _ int, c *tsdb.Column) { putSeries(t, c, points, 1) },
			want:    func(int) int64 { return points + 1 }},
		{name: "every second column written again since", idleAfter: idle, wait: idle + time.Millisecond,
			between: func(t *testing.T, i int, c *tsdb.Column) {
				if i%2 == 1 {
					putSeries(t, c, points, 1)
				}
			},
			want: func(i int) int64 {
				if i%2 == 1 {
					return points + 1
				}
				return 0
			}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			clk := newClock()
			t0 := clk.now()
			dir := t.TempDir()
			opts := tsdb.Options{IdleAfter: tt.idleAfter, Now: clk.now}
			db := openDB(t, dir, opts)
			before := liveHeap()
			held := make([]*tsdb.Column, cols)
			for i := range held {
				held[i] = column(t, db, "b", seriesName(i), tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				putSeries(t, held[i], 0, points)
			}
			if lo, hi := unsyncedOf(held); lo != points || hi != points {
				t.Fatalf("premise: Unsynced %d to %d, want %d", lo, hi, points)
			}
			clk.advance(30 * time.Minute)
			if tt.between != nil {
				for i, c := range held {
					tt.between(t, i, c)
				}
			}
			clk.advance(tt.wait - 30*time.Minute)
			must(t, db.ReleaseIdle())
			per := (liveHeap() - before) / int64(cols)

			for i, c := range held {
				if got := c.Stats().Unsynced; got != tt.want(i) {
					t.Errorf("column %d after %s: Unsynced %d, want %d", i, clk.since(t0), got, tt.want(i))
					break
				}
			}
			if tt.memory > 0 {
				t.Logf("%d B per released column, at most %d", per, tt.memory)
				if per > tt.memory {
					t.Errorf("%d B per released column, at most %d", per, tt.memory)
				}
			}
			runtime.KeepAlive(held)

			// what ReleaseIdle secured survives a crash
			released := 0
			for i := range cols {
				if tt.want(i) == 0 {
					released++
				}
			}
			if released > 0 {
				crashed := openDB(t, crashCopy(t, dir), opts)
				for i := range cols {
					n, err := lookup(t, crashed, "b", seriesName(i)).Count()
					must(t, err)
					if tt.want(i) == 0 && n != points {
						t.Errorf("column %d: %d points after a crash, want the %d released ones", i, n, points)
						break
					}
				}
			}

			// the held handles stay valid and are the ones the database hands out
			for i, c := range held {
				again, err := db.Column("b", seriesName(i), tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				must(t, err)
				if again != c {
					t.Fatalf("column %d: DB.Column returns another *Column after ReleaseIdle", i)
				}
				putSeries(t, c, points+1, 10)
				must(t, c.Sync())
				n, err := c.Count()
				must(t, err)
				want := int64(points + 10)
				if tt.want(i) > points {
					want++
				}
				if n != want {
					t.Fatalf("column %d: %d points after 10 more and a Sync, want %d", i, n, want)
				}
			}
		})
	}
}
