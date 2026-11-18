// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The acceptance tests of R-DEC-TS-NAGO-CURSOR (wokoda N23, L-N23-18, L-N23-26, L-N23-32, L-N23-33) and of the row of
// the cursor of R-DEC-TS-NAGO-MEMORY: a call of CursorF64.Next yields exactly what a fresh ScanF64 of the same range
// yields at this moment, also after any exported method of Column and DB between two calls. The order of a row is: the
// first call, the method, the second call, only then the fresh ScanF64; a fresh ScanF64 or Stats before the second
// call would take the lock of the column and let it start anew, even if the method in between forgot the counter
// (AGENTS §8, "Das Messinstrument ist selbst falsch"). That the cursor continues the test sees in the bytes the
// FilePool reads, without a counter only for tests.

// n23Scan is what one call of a scan gave fn: its batches, the bits of their values, and its error.
type n23Scan struct {
	ts   [][]int64
	bits [][]uint64
	err  string
}

func (s n23Scan) String() string {
	n := 0
	for _, b := range s.ts {
		n += len(b)
	}
	first, last := int64(0), int64(0)
	if n > 0 {
		first, last = s.ts[0][0], s.ts[len(s.ts)-1][len(s.ts[len(s.ts)-1])-1]
	}
	return fmt.Sprintf("%d batches, %d points %d … %d, err %q", len(s.ts), n, first, last, s.err)
}

func (s n23Scan) equal(o n23Scan) bool {
	if s.err != o.err || len(s.ts) != len(o.ts) {
		return false
	}
	for i := range s.ts {
		if !slices.Equal(s.ts[i], o.ts[i]) || !slices.Equal(s.bits[i], o.bits[i]) {
			return false
		}
	}
	return true
}

// last is the time of the last point the call gave fn; ok is false without one.
func (s n23Scan) last() (int64, bool) {
	if len(s.ts) == 0 {
		return 0, false
	}
	b := s.ts[len(s.ts)-1]
	return b[len(b)-1], true
}

// n23Collect returns fn for a scan, which keeps the batches in s and stops after stop batches (0: never).
func n23Collect(s *n23Scan, stop int) func(ts []int64, vals []float64) bool {
	return func(ts []int64, vals []float64) bool {
		bits := make([]uint64, len(vals))
		for i, v := range vals {
			bits[i] = math.Float64bits(v)
		}
		s.ts, s.bits = append(s.ts, slices.Clone(ts)), append(s.bits, bits)
		return stop == 0 || len(s.ts) < stop
	}
}

func n23Fresh(c *tsdb.Column, min, max int64, stop int) n23Scan {
	var s n23Scan
	if err := c.ScanF64(min, max, n23Collect(&s, stop)); err != nil {
		s.err = err.Error()
	}
	return s
}

func n23Next(cu *tsdb.CursorF64, min, max int64, stop int) n23Scan {
	var s n23Scan
	if err := cu.Next(min, max, n23Collect(&s, stop)); err != nil {
		s.err = err.Error()
	}
	return s
}

// n23CurEnv is the database of a row: a decimal column with chunks of 40 points and blocks of 8, of the points 1 … 100
// a second apart (value i/10), Close and an open again, so that they are on disk.
type n23CurEnv struct {
	t     *testing.T
	dir   string
	opts  tsdb.Options
	clock *clock
	db    *tsdb.DB
	c     *tsdb.Column
}

const (
	n23CurBucket = "cur"
	n23CurCol    = "c"
	n23CurMax    = int64(10_000_000) // the end of the range of both calls
)

func n23CurOptions(clk *clock) tsdb.Options {
	return tsdb.Options{BlockPoints: 8, Split: tsdb.SplitByCount(40), CompactTombstoneRatio: 2, FlushBytes: 1 << 30,
		IdleAfter: time.Minute, Now: clk.now}
}

func newN23CurEnv(t *testing.T, opt func(o *tsdb.Options)) *n23CurEnv {
	t.Helper()
	e := &n23CurEnv{t: t, dir: t.TempDir(), clock: newClock()}
	e.opts = n23CurOptions(e.clock)
	db := openDB(t, e.dir, e.opts)
	c := column(t, db, n23CurBucket, n23CurCol, tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1})
	for i := int64(1); i <= 100; i++ {
		must(t, c.PutI64(i*1000, i))
	}
	closeDB(t, db)
	if opt != nil {
		opt(&e.opts)
	}
	e.db = openDB(t, e.dir, e.opts)
	e.c = lookup(t, e.db, n23CurBucket, n23CurCol)
	return e
}

func (e *n23CurEnv) colDir() string { return filepath.Join(e.dir, n23CurBucket, n23CurCol) }

// n23CurRow is a row of the table: a method of Column or DB (by its name in the method set) and what the row does
// with it between the two calls of the cursor.
type n23CurRow struct {
	method string // "Column.<name>" or "DB.<name>"
	name   string
	opt    func(o *tsdb.Options)
	act    func(e *n23CurEnv)
}

// n23WaitCompaction waits, without the lock of the column, until a compaction in the background has put its chunks in
// place: a chunk file differs from before (os.ReadDir and os.ReadFile) and no file of a compaction in progress (.new,
// compact.log) is left. It fails loudly, if that does not happen within 10 s.
func n23WaitCompaction(e *n23CurEnv, before snapshot) {
	e.t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		entries, err := os.ReadDir(e.colDir())
		must(e.t, err)
		busy := false
		for _, en := range entries {
			if n := en.Name(); strings.HasSuffix(n, ".new") || strings.HasPrefix(n, "compact.log") {
				busy = true
			}
		}
		if !busy {
			if now, err := n23TryReadSnapshot(e.colDir()); err == nil && n22ChunksChanged(before, now) {
				return
			}
		}
		if time.Now().After(deadline) {
			e.t.Fatalf("no compaction in the background put its chunks in place: %v", entries)
		}
		runtime.Gosched()
	}
}

// n23TryReadSnapshot is readSnapshot, which returns an error instead of failing, because a compaction may remove a
// file between the listing and the read.
func n23TryReadSnapshot(dir string) (snapshot, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	s := snapshot{}
	for _, en := range entries {
		b, err := os.ReadFile(filepath.Join(dir, en.Name()))
		if err != nil {
			return nil, err
		}
		s[en.Name()] = b
	}
	return s, nil
}

func mapsEqualBool(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func n23ChunkNames(e *n23CurEnv) map[string]bool {
	e.t.Helper()
	m, _ := filepath.Glob(filepath.Join(e.colDir(), "*.tsb"))
	names := map[string]bool{}
	for _, p := range m {
		names[filepath.Base(p)] = true
	}
	return names
}

// n23CurRows are the rows: every exported method of Column and DB, and the rows of the table of the contract.
func n23CurRows() []n23CurRow {
	return []n23CurRow{
		{method: "Column.Count", name: "Count", act: func(e *n23CurEnv) { _, _ = e.c.Count() }},
		{method: "Column.Delete", name: "Delete in the range", act: func(e *n23CurEnv) { must(e.t, e.c.Delete(30_000)) }},
		{method: "Column.DeleteRange", name: "DeleteRange in the range", act: func(e *n23CurEnv) { must(e.t, e.c.DeleteRange(30_000, 35_000)) }},
		{method: "Column.DropBefore", name: "DropBefore, which deletes the chunk of the place", act: func(e *n23CurEnv) {
			if _, err := e.c.DropBefore(41_000); err != nil {
				e.t.Fatal(err)
			}
		}},
		{method: "Column.Flush", name: "Flush", act: func(e *n23CurEnv) { must(e.t, e.c.Flush()) }},
		{method: "Column.IterF64", name: "IterF64", act: func(e *n23CurEnv) {
			for range e.c.IterF64(minTS, maxTS) {
			}
		}},
		{method: "Column.IterI64", name: "IterI64", act: func(e *n23CurEnv) {
			for range e.c.IterI64(minTS, maxTS) {
			}
		}},
		{method: "Column.IterString", name: "IterString", act: func(e *n23CurEnv) {
			for range e.c.IterString(minTS, maxTS) {
			}
		}},
		{method: "Column.NewCursorF64", name: "a second cursor on the same column reads", act: func(e *n23CurEnv) {
			cu := e.c.NewCursorF64()
			defer cu.Close()
			_ = n23Next(cu, minTS, maxTS, 0)
		}},
		{method: "Column.PutF64", name: "PutF64 behind the last point", act: func(e *n23CurEnv) { must(e.t, e.c.PutF64(101_000, 10.1)) }},
		{method: "Column.PutI64", name: "Put behind the last point", act: func(e *n23CurEnv) { must(e.t, e.c.PutI64(101_000, 101)) }},
		{method: "Column.PutI64", name: "Put behind the last point over a change of chunk of the split policy", act: func(e *n23CurEnv) {
			before := n23ChunkNames(e)
			for ts := int64(101); ts <= 125; ts++ {
				must(e.t, e.c.PutI64(ts*1000, ts))
			}
			if mapsEqualBool(before, n23ChunkNames(e)) {
				e.t.Fatalf("premise: no new chunk after the split policy")
			}
		}},
		{method: "Column.PutI64", name: "Put out of order in the range of the second call", act: func(e *n23CurEnv) { must(e.t, e.c.PutI64(50_000, 777)) }},
		{method: "Column.PutString", name: "PutString (fails on a decimal column)", act: func(e *n23CurEnv) { _ = e.c.PutString(102_000, "x") }},
		{method: "Column.ScanF64", name: "ScanF64", act: func(e *n23CurEnv) { _ = n23Fresh(e.c, minTS, maxTS, 0) }},
		{method: "Column.ScanI64", name: "ScanI64", act: func(e *n23CurEnv) {
			_ = e.c.ScanI64(minTS, maxTS, func([]int64, []int64) bool { return true })
		}},
		{method: "Column.ScanString", name: "ScanString", act: func(e *n23CurEnv) {
			_ = e.c.ScanString(minTS, maxTS, func([]int64, []string) bool { return true })
		}},
		{method: "Column.Schema", name: "Schema", act: func(e *n23CurEnv) { _ = e.c.Schema() }},
		{method: "Column.Stats", name: "Stats", act: func(e *n23CurEnv) { _ = e.c.Stats() }},
		{method: "Column.Sync", name: "Sync after a point behind the last", act: func(e *n23CurEnv) {
			must(e.t, e.c.PutI64(101_000, 101))
			must(e.t, e.c.Sync())
		}},
		{method: "", name: "a compaction, synchronous at MaxHeadPoints", opt: func(o *tsdb.Options) { o.MaxHeadPoints = 2 },
			act: func(e *n23CurEnv) {
				before := n23ChunkNames(e)
				snap := readSnapshot(e.t, e.colDir())
				for _, ts := range []int64{20_000, 50_000, 60_000} {
					must(e.t, e.c.PutI64(ts, ts/10))
				}
				if !n22ChunksChanged(snap, readSnapshot(e.t, e.colDir())) {
					e.t.Fatalf("premise: no compaction at MaxHeadPoints (chunks %v)", before)
				}
			}},
		{method: "", name: "a compaction in the background after FlushBytes", opt: func(o *tsdb.Options) { o.FlushBytes = 16 },
			act: func(e *n23CurEnv) {
				before := readSnapshot(e.t, e.colDir())
				// corrections of more than FlushBytes of payload
				for ts := int64(20_000); ts <= 70_000; ts += 5_000 {
					must(e.t, e.c.PutI64(ts, ts/10))
				}
				n23WaitCompaction(e, before)
			}},
		{method: "DB.Close", name: "DB.Close", act: func(e *n23CurEnv) { must(e.t, e.db.Close()) }},
		{method: "DB.Column", name: "DB.Column of the same column", act: func(e *n23CurEnv) {
			_, err := e.db.Column(n23CurBucket, n23CurCol, tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1})
			must(e.t, err)
		}},
		{method: "DB.DropBucket", name: "DropBucket", act: func(e *n23CurEnv) { must(e.t, e.db.DropBucket(n23CurBucket)) }},
		{method: "DB.DropColumn", name: "DropColumn", act: func(e *n23CurEnv) { must(e.t, e.db.DropColumn(n23CurBucket, n23CurCol)) }},
		{method: "DB.LookupColumn", name: "DB.LookupColumn", act: func(e *n23CurEnv) { _, _, _ = e.db.LookupColumn(n23CurBucket, n23CurCol) }},
		{method: "DB.ReleaseIdle", name: "ReleaseIdle after IdleAfter, over a point only in memory", act: func(e *n23CurEnv) {
			must(e.t, e.c.PutI64(101_000, 101))
			snap := readSnapshot(e.t, e.colDir()) // with os: Stats would take the lock of the column
			e.clock.advance(2 * time.Minute)
			must(e.t, e.db.ReleaseIdle())
			if !n22ChunksChanged(snap, readSnapshot(e.t, e.colDir())) {
				e.t.Fatalf("premise: ReleaseIdle saved nothing")
			}
		}},
		{method: "DB.SeriesColumns", name: "SeriesColumns", act: func(e *n23CurEnv) { _, _ = e.db.SeriesColumns() }},
	}
}

// TestN23CursorMethodTable: the table of TestN23CursorInterleaved holds every exported method of *tsdb.Column and
// *tsdb.DB, found by reflect, and no method that does not exist; a new method fails it, until the table interleaves it.
func TestN23CursorMethodTable(t *testing.T) {
	rows := map[string]bool{}
	for _, r := range n23CurRows() {
		if r.method != "" {
			rows[r.method] = true
		}
	}
	methods := map[string]bool{}
	for prefix, typ := range map[string]reflect.Type{"Column": reflect.TypeOf(&tsdb.Column{}), "DB": reflect.TypeOf(&tsdb.DB{})} {
		for i := range typ.NumMethod() {
			methods[prefix+"."+typ.Method(i).Name] = true
		}
	}
	for m := range methods {
		if !rows[m] {
			t.Errorf("the exported method %s has no row: the cursor test does not interleave it", m)
		}
	}
	for m := range rows {
		if !methods[m] {
			t.Errorf("the row %s names no exported method", m)
		}
	}
	if len(methods) < 20 {
		t.Fatalf("premise: reflect found %d methods", len(methods))
	}
}

// TestN23CursorInterleaved: the first call of the cursor over the whole range stops after its first batch; then the
// method of the row; then the second call from the time after the last point the first gave fn; only then a fresh
// ScanF64 of the range of the second call. Both calls yield what the fresh ScanF64 yields, batch for batch and bit for
// bit, its error included (after DB.Close, DropBucket and DropColumn the one of a fresh ScanF64 of the same Column).
// The counter row "nothing in between" is the row of every row.
func TestN23CursorInterleaved(t *testing.T) {
	rows := append([]n23CurRow{{name: "counter row: nothing between the calls", act: func(*n23CurEnv) {}}}, n23CurRows()...)
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			e := newN23CurEnv(t, r.opt)
			want1 := n23Fresh(e.c, minTS, n23CurMax, 1) // before the first call, which starts anew anyway
			cu := e.c.NewCursorF64()
			defer cu.Close()
			got1 := n23Next(cu, minTS, n23CurMax, 1)
			if !got1.equal(want1) {
				t.Fatalf("the first call yields %v, a fresh ScanF64 %v", got1, want1)
			}
			last, ok := got1.last()
			if !ok {
				t.Fatalf("premise: the first call gave fn no point")
			}
			r.act(e)
			got2 := n23Next(cu, last+1, n23CurMax, 0)
			want2 := n23Fresh(e.c, last+1, n23CurMax, 0)
			if !got2.equal(want2) {
				t.Errorf("the second call yields %v, a fresh ScanF64 after it %v", got2, want2)
			}
			if r.method == "DB.Close" || r.method == "DB.DropBucket" || r.method == "DB.DropColumn" {
				return
			}
			// a third call goes on from the second; it also yields what a fresh ScanF64 yields
			got3 := n23Next(cu, minTS, n23CurMax, 0)
			if want3 := n23Fresh(e.c, minTS, n23CurMax, 0); !got3.equal(want3) {
				t.Errorf("a third call over the whole range yields %v, a fresh ScanF64 %v", got3, want3)
			}
		})
	}
}

// TestN23CursorOtherScheme: NewCursorF64 of a column of another scheme fails at the first Next with the error of
// ScanF64.
func TestN23CursorOtherScheme(t *testing.T) {
	for _, sc := range []tsdb.Scheme{tsdb.SchemeEnum, tsdb.SchemeString} {
		t.Run(sc.String(), func(t *testing.T) {
			db := openDB(t, t.TempDir(), tsdb.Options{})
			c := column(t, db, "cur", "s", tsdb.Schema{Scheme: sc})
			must(t, c.PutString(1000, "a"))
			want := n23Fresh(c, minTS, maxTS, 0)
			if want.err == "" {
				t.Fatalf("premise: ScanF64 of a %s column does not fail", sc)
			}
			cu := c.NewCursorF64()
			defer cu.Close()
			if got := n23Next(cu, minTS, maxTS, 0); !got.equal(want) {
				t.Errorf("the first Next yields %v, ScanF64 %v", got, want)
			}
		})
	}
}

// TestN23CursorContinuesByReadBytes: ten calls of a cursor, each from the time after the last point of the one before
// and stopping after its first batch, read fewer bytes through the FilePool without a change between the calls than
// with a point written between each two (behind the newest point, outside of their range), which lets each call start
// anew. Both yield, all ten together, what a fresh ScanF64 of their range yields afterwards.
func TestN23CursorContinuesByReadBytes(t *testing.T) {
	run := func(t *testing.T, change bool) int64 {
		dir := t.TempDir()
		pool := ndb.NewFilePool(64)
		opts := tsdb.Options{BlockPoints: 8, FilePool: pool}
		db := openDB(t, dir, opts)
		c := column(t, db, "cur", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
		r := rand.New(rand.NewPCG(23, 18))
		for i := int64(1); i <= 400; i++ {
			must(t, c.PutI64(i*1000, r.Int64N(1_000_000)))
		}
		closeDB(t, db)
		db = openDB(t, dir, opts)
		c = lookup(t, db, "cur", "c")
		cu := c.NewCursorF64()
		defer cu.Close()
		var read int64
		from := int64(minTS)
		next := int64(1_000_000)
		var all n23Scan
		for k := range 10 {
			if change && k > 0 {
				must(t, c.PutI64(next, 1))
				next += 1000
			}
			r0 := pool.Stats().ReadBytes
			got := n23Next(cu, from, 400_000, 1)
			read += pool.Stats().ReadBytes - r0
			last, ok := got.last()
			if !ok || got.err != "" {
				t.Fatalf("call %d: %v", k+1, got)
			}
			all.ts, all.bits = append(all.ts, got.ts...), append(all.bits, got.bits...)
			from = last + 1
		}
		if want := n23Fresh(c, minTS, 400_000, 10); !all.equal(want) {
			t.Errorf("the ten calls yield %v, a fresh ScanF64 of their range %v", all, want)
		}
		closeDB(t, db)
		return read
	}
	quiet := run(t, false)
	changed := run(t, true)
	t.Logf("read bytes of ten calls: %d without a change, %d with a change between each two", quiet, changed)
	if quiet >= changed {
		t.Errorf("the calls read %d bytes without a change and %d with one: the cursor does not continue", quiet, changed)
	}
}

// TestN23CursorOpenBlockLast: a column with a finalized chunk of two blocks and an open chunk, which holds continuation
// records only (after Sync), together at least twice as large as the finalized chunk. A first call over the whole
// range, whose fn stops in the first block; a call from the time after its last point, whose fn stops in the next block
// (it continues); a call over the whole range again, which starts anew, whose fn stops in the first block: each reads
// fewer bytes through the FilePool than the records together. Counter row: the first call of a new cursor over the
// same range with an fn that does not stop reads at least as many bytes as the records together.
func TestN23CursorOpenBlockLast(t *testing.T) {
	dir := t.TempDir()
	pool := ndb.NewFilePool(64)
	db := openDB(t, dir, tsdb.Options{BlockPoints: 8, FilePool: pool})
	c := column(t, db, "cur", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
	for i := int64(1); i <= 16; i++ {
		must(t, c.PutI64(i*1000, i))
	}
	must(t, c.Flush())
	closeDB(t, db)
	final, _ := filepath.Glob(filepath.Join(dir, "cur", "c", "*_?*.tsb"))
	if len(final) != 1 {
		t.Fatalf("premise: %d finalized chunks", len(final))
	}
	finalSize := n23Size(t, final[0])

	db = openDB(t, dir, tsdb.Options{BlockPoints: 4096, FilePool: pool})
	c = lookup(t, db, "cur", "c")
	r := rand.New(rand.NewPCG(23, 32))
	ts := int64(100_000)
	var header int64
	// 800 points in 40 records: fewer than a block, so that the open block is never sealed
	for range 40 {
		for range 20 {
			must(t, c.PutI64(ts, r.Int64N(1<<40)))
			ts += 1000
			if header == 0 {
				open, _ := filepath.Glob(filepath.Join(dir, "cur", "c", "*_.tsb"))
				if len(open) != 1 {
					t.Fatalf("premise: %d open chunks after the first point", len(open))
				}
				header = n23Size(t, open[0])
			}
		}
		must(t, c.Sync())
	}
	open, _ := filepath.Glob(filepath.Join(dir, "cur", "c", "*_.tsb"))
	if len(open) != 1 {
		t.Fatalf("premise: %d open chunks", len(open))
	}
	records := n23Size(t, open[0]) - header // the records behind the header of the open chunk
	if records < 2*finalSize {
		t.Fatalf("premise: the records have %d bytes, the finalized chunk %d: not twice as large", records, finalSize)
	}
	measure := func(cu *tsdb.CursorF64, min int64, stop int) (n23Scan, int64) {
		r0 := pool.Stats().ReadBytes
		s := n23Next(cu, min, maxTS, stop)
		return s, pool.Stats().ReadBytes - r0
	}
	cu := c.NewCursorF64()
	defer cu.Close()
	first, r1 := measure(cu, minTS, 1)
	last, ok := first.last()
	if !ok || last > 8000 {
		t.Fatalf("premise: the first call did not stop in the first block (%v)", first)
	}
	second, r2 := measure(cu, last+1, 1)
	if l, ok := second.last(); !ok || l > 16000 {
		t.Fatalf("premise: the second call did not stop in the next block (%v)", second)
	}
	_, r3 := measure(cu, minTS, 1)
	for i, rb := range []int64{r1, r2, r3} {
		if rb >= records {
			t.Errorf("call %d (%s) read %d bytes, the records have %d: it loaded the open block before the sealed blocks were through",
				i+1, []string{"first", "continued", "anew"}[i], rb, records)
		}
	}
	// counter row: the bytes read see the load of the records
	cu2 := c.NewCursorF64()
	defer cu2.Close()
	all, r4 := measure(cu2, minTS, 0)
	if r4 < records {
		t.Errorf("counter row: a call over all points read %d bytes, the records have %d: the measure does not see their load", r4, records)
	}
	if want := n23Fresh(c, minTS, maxTS, 0); !all.equal(want) {
		t.Errorf("counter row: the call yields %v, a fresh ScanF64 %v", all, want)
	}
	t.Logf("read: %d, %d, %d bytes; all %d; the records %d, the finalized chunk %d", r1, r2, r3, r4, records, finalSize)
}

// TestN23CursorMemory (R-DEC-TS-NAGO-MEMORY, the row of the cursor): after Close a cursor, which the test holds beyond
// Close, so that not only the garbage collector frees its buffer, keeps at most 4 KiB for reading per column. Counter
// row: the same cursors between two calls of Next, before Close, hold more than 4 KiB per column: the measure sees the
// buffer. 100 columns of three sealed blocks of 4096 points (the default), with -short a quarter of them.
func TestN23CursorMemory(t *testing.T) {
	cols := columns(100)
	dir := t.TempDir()
	db := openDB(t, dir, tsdb.Options{})
	const points = 3*4096 + 100
	var cs []*tsdb.Column
	for i := range cols {
		c := column(t, db, "b", seriesName(i), tsdb.Schema{Scheme: tsdb.SchemeDecimal})
		for p := range points {
			must(t, c.PutI64(int64(p)*10, int64(p*37%1001)))
		}
		cs = append(cs, c)
	}
	closeDB(t, db)
	db = openDB(t, dir, tsdb.Options{})
	cs = cs[:0]
	for i := range cols {
		c := lookup(t, db, "b", seriesName(i))
		// a query before all points opens the column without a block
		must(t, c.ScanF64(-1000, -1, func([]int64, []float64) bool { return true }))
		cs = append(cs, c)
	}
	before := liveHeap()
	curs := make([]*tsdb.CursorF64, cols)
	for i, c := range cs {
		curs[i] = c.NewCursorF64()
		if s := n23Next(curs[i], minTS, maxTS, 1); len(s.ts) != 1 || s.err != "" {
			t.Fatalf("premise: the first call yields %v", s)
		}
	}
	between := (liveHeap() - before) / int64(cols)
	for _, cu := range curs {
		cu.Close()
	}
	after := (liveHeap() - before) / int64(cols)
	runtime.KeepAlive(curs)
	runtime.KeepAlive(cs)
	t.Logf("per column: %d B between two calls of Next, %d B after Close", between, after)
	if between <= 4*kib {
		t.Errorf("counter row: %d B per column between two calls of Next, want more than %d: the measure does not see the buffer", between, 4*kib)
	}
	if after > 4*kib {
		t.Errorf("%d B per column for reading after Close of the cursor, at most %d", after, 4*kib)
	}
}
