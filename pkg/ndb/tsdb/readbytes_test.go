// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"fmt"
	"hash/fnv"
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/tsdb"
)

const mib = 1024 * kib

// chunkFile is a chunk on disk with the time range of its name.
type chunkFile struct {
	name     string
	min, max int64 // max is -1 for the open chunk
	size     int64
}

func chunkFiles(t testing.TB, colDir string) []chunkFile {
	t.Helper()
	entries, err := os.ReadDir(colDir)
	must(t, err)
	var res []chunkFile
	for _, e := range entries {
		name := e.Name()
		if !chunkName.MatchString(name) && !openName.MatchString(name) {
			continue
		}
		lo, hi, _ := strings.Cut(strings.TrimSuffix(name, ".tsb"), "_")
		cf := chunkFile{name: name, max: -1}
		cf.min, _ = strconv.ParseInt(lo, 10, 64)
		if hi != "" {
			cf.max, _ = strconv.ParseInt(hi, 10, 64)
		}
		fi, err := e.Info()
		must(t, err)
		cf.size = fi.Size()
		res = append(res, cf)
	}
	slices.SortFunc(res, func(a, b chunkFile) int { return int(a.min - b.min) })
	return res
}

// statsDelta is how much a pool counted while fn ran.
func statsDelta(pool *ndb.FilePool, fn func()) ndb.FilePoolStats {
	before := pool.Stats()
	fn()
	after := pool.Stats()
	return ndb.FilePoolStats{
		Opens:        after.Opens - before.Opens,
		ReadBytes:    after.ReadBytes - before.ReadBytes,
		WrittenBytes: after.WrittenBytes - before.WrittenBytes,
	}
}

// bigColumn writes n points of 50 Hz with random values, which take about 6 B each without compression, into the column
// b/c of dir and closes it. Point i has the time i*20.
func bigColumn(t testing.TB, dir string, opts tsdb.Options, n int) {
	t.Helper()
	db := openDB(t, dir, opts)
	c := column(t, db, "b", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
	rng := rand.New(rand.NewPCG(1, 2))
	for i := range n {
		must(t, c.PutI64(int64(i)*20, rng.Int64N(1<<40)))
	}
	closeDB(t, db)
}

func countRange(t testing.TB, c *tsdb.Column, min, max int64) int64 {
	t.Helper()
	var n int64
	must(t, c.ScanI64(min, max, func(ts, _ []int64) bool { n += int64(len(ts)); return true }))
	return n
}

// TestQueryReadsOnlyTheBlocksItNeeds: a query reads from a chunk only the blocks that overlap its range, and at most 2 MiB
// besides, also at the first access after the open, counted in the bytes the FilePool read. The column has three chunks
// of about 4 MiB in blocks of 4096 points of at most 48 KiB; a minute of 50 Hz is 3000 points and overlaps at most two
// blocks. The counter row reads all points and shows that the pool counts what the column reads: every block.
func TestQueryReadsOnlyTheBlocksItNeeds(t *testing.T) {
	const (
		points   = 2_100_000
		minute   = 60_000
		maxBlock = 48 * kib
	)
	opts := tsdb.Options{Split: tsdb.SplitBySize(4 * mib)}
	dir := t.TempDir()
	bigColumn(t, dir, opts, points)
	chunks := chunkFiles(t, filepath.Join(dir, "b", "c"))
	if len(chunks) != 3 {
		t.Fatalf("premise: %d chunks %v, want 3", len(chunks), chunks)
	}
	var total int64
	for _, cf := range chunks {
		total += cf.size
	}
	middle := (chunks[1].min + chunks[1].max) / 2
	last := int64(points-1) * 20

	for _, tt := range []struct {
		name     string
		warm     [2]int64 // a query before the measured one, if not zero
		min, max int64
		atMost   int64 // -1 for no upper bound
		atLeast  int64
	}{
		{name: "the first access: a minute in the middle of a chunk", min: middle, max: middle + minute - 1, atMost: 2*mib + 2*maxBlock},
		{name: "the first access: a minute across two chunks", min: chunks[1].max - minute/2, max: chunks[1].max + minute/2 - 1, atMost: 2*mib + 2*maxBlock},
		{name: "the first access: the newest minute in the open chunk", min: last - minute + 20, max: last, atMost: 2*mib + 2*maxBlock},
		{name: "the first access: a range before all points", min: -minute, max: -1, atMost: 2 * mib},
		{name: "a minute after a minute in another chunk", warm: [2]int64{chunks[0].min + minute, chunks[0].min + 2*minute - 1},
			min: middle, max: middle + minute - 1, atMost: 2*mib + 2*maxBlock},
		{name: "counter row: all points", min: minTS, max: maxTS, atMost: -1, atLeast: total - 64*kib},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pool := ndb.NewFilePool(1024)
			t.Cleanup(func() { _ = pool.Close() })
			o := opts
			o.FilePool = pool
			db := openDB(t, dir, o)
			var c *tsdb.Column
			if tt.warm != [2]int64{} {
				c = lookup(t, db, "b", "c")
				countRange(t, c, tt.warm[0], tt.warm[1])
			}
			var n int64
			d := statsDelta(pool, func() {
				if c == nil {
					c = lookup(t, db, "b", "c")
				}
				n = countRange(t, c, tt.min, tt.max)
			})
			want := int64(0)
			for i := range int64(points) {
				if ts := i * 20; ts >= tt.min && ts <= tt.max {
					want++
				}
			}
			if n != want {
				t.Fatalf("premise: %d points in the range, want %d", n, want)
			}
			t.Logf("%d points of %d B of chunks: %d B read, %d files opened", n, total, d.ReadBytes, d.Opens)
			if tt.atMost >= 0 && d.ReadBytes > tt.atMost {
				t.Errorf("the query read %d B, at most %d", d.ReadBytes, tt.atMost)
			}
			if d.ReadBytes < tt.atLeast {
				t.Errorf("the query read %d B, at least %d", d.ReadBytes, tt.atLeast)
			}
			closeDB(t, db)
		})
	}
}

// openFixture is a column in a state after Close or after a crash, for the tests of the open.
type openFixture struct {
	dir      string
	schema   tsdb.Schema
	base     []string         // the value of the point at the time i*20, all of them secured
	extra    map[int64]string // corrections and later points
	optional map[int64]bool   // of extra: those of the torn last write, which may be missing
	colFiles int              // chunk files on disk
}

// scan reads every point of the column, checks it against the fixture and returns the number of points and a sum over
// them; skip leaves one time out of the sum.
func (f openFixture) scan(t testing.TB, s surfaceScheme, c *tsdb.Column, skip int64) (int, uint64) {
	t.Helper()
	h := fnv.New64a()
	n, required, wantRequired := 0, 0, len(f.base)
	for ts := range f.extra {
		if !f.optional[ts] && (ts%20 != 0 || ts/20 >= int64(len(f.base))) {
			wantRequired++
		}
	}
	last := int64(minTS)
	check := func(ts int64, v string) {
		if ts <= last && n > 0 {
			t.Fatalf("the time %d follows %d", ts, last)
		}
		last = ts
		want, ok := f.extra[ts]
		if !ok && ts%20 == 0 && ts/20 < int64(len(f.base)) {
			want, ok = f.base[ts/20], true
		}
		if !ok || v != want {
			t.Fatalf("the time %d reads %q, which was never written there (want %q)", ts, v, want)
		}
		n++
		if !f.optional[ts] {
			required++
		}
		if ts != skip {
			_, _ = fmt.Fprintf(h, "%d=%s;", ts, v)
		}
	}
	if s.schema.Scheme == tsdb.SchemeDecimal {
		must(t, c.ScanI64(minTS, maxTS, func(ts, vs []int64) bool {
			for i := range ts {
				check(ts[i], strconv.FormatInt(vs[i], 10))
			}
			return true
		}))
	} else {
		must(t, c.ScanString(minTS, maxTS, func(ts []int64, vs []string) bool {
			for i := range ts {
				check(ts[i], vs[i])
			}
			return true
		}))
	}
	if required != wantRequired {
		t.Fatalf("%d of the %d secured points read", required, wantRequired)
	}
	return n, h.Sum64()
}

// makeOpenFixture writes ten small finalized chunks and an open chunk of more than 2 MiB into b/c, for enum with -short
// one of less. With crash it then writes corrections in the first chunk into the head, a Sync of 7 more points, and
// leaves a crash whose last write is torn by 3 bytes, so that the open takes the open chunk up to its last intact state.
func makeOpenFixture(t testing.TB, s surfaceScheme, crash bool) openFixture {
	t.Helper()
	const small = 5000
	points, big := 360_000, true // about 9 B each for decimal
	if s.schema.Scheme == tsdb.SchemeEnum {
		points = 1_400_000 // about 2.5 B each for ids far apart
		if testing.Short() {
			points, big = 120_000, false
		}
	}
	split := func(st tsdb.ChunkStats) bool {
		return st.NextMillis < 10*small*20 && st.NextMillis/(small*20) != st.MinMillis/(small*20)
	}
	opts := tsdb.Options{Split: split}
	dir := t.TempDir()
	db := openDB(t, dir, opts)
	c := column(t, db, "b", "c", s.schema)
	f := openFixture{dir: dir, schema: s.schema, base: make([]string, points), extra: map[int64]string{}, optional: map[int64]bool{}}
	rng := rand.New(rand.NewPCG(3, 4))
	val := func() string {
		if s.schema.Scheme == tsdb.SchemeEnum {
			return "n" + strconv.Itoa(rng.IntN(1<<14)) // ids far apart, so that the open chunk grows beyond 2 MiB
		}
		return strconv.FormatInt(rng.Int64(), 10)
	}
	for i := range points {
		f.base[i] = val()
		must(t, s.put(c, int64(i)*20, f.base[i]))
	}
	closeDB(t, db)
	if crash {
		db = openDB(t, dir, opts)
		c = lookup(t, db, "b", "c")
		put := func(ts int64, optional bool) {
			v := val()
			must(t, s.put(c, ts, v))
			f.extra[ts] = v
			f.optional[ts] = optional
		}
		for i := range 20 {
			put(int64(i*100)*20+20, false) // corrections in the first chunk, into the head
		}
		put(int64(points)*20, false)
		must(t, c.Sync())
		for i := range 7 {
			put(int64(points+1+i)*20, true)
		}
		must(t, c.Sync())
		crashed := crashCopy(t, dir)
		closeDB(t, db)
		open := chunkFiles(t, filepath.Join(crashed, "b", "c"))
		path := filepath.Join(crashed, "b", "c", open[len(open)-1].name)
		if open[len(open)-1].max != -1 {
			t.Fatalf("premise: no open chunk %v", open)
		}
		must(t, os.Truncate(path, open[len(open)-1].size-3))
		f.dir = crashed
	}
	files := chunkFiles(t, filepath.Join(f.dir, "b", "c"))
	f.colFiles = len(files)
	if f.colFiles < 10 || files[len(files)-1].max != -1 || big && files[len(files)-1].size < 2*mib+256*kib {
		t.Fatalf("premise: chunk files %v, want 10 finalized and an open one of more than 2 MiB", files)
	}
	return f
}

// TestOpeningAColumnReadsLittle: the open of a column opens at most schema.json and its open chunk and reads at most 2 MiB
// of it, counted by the FilePool; other chunks, the dictionary and the head are read when a query or a write needs them.
// After a crash the open chunk is still taken up to its last intact state. The column has ten finalized chunks and an
// open chunk of more than 2 MiB, after a crash also corrections in the head and a torn last write. The counter row queries
// all points after the open: the open and the query together open every chunk and read every byte of them, which shows
// that the pool sees what the column reads. Then the column goes on and survives another crash; with -short only after
// the open by DB.Column. With -short the open chunk of enum is smaller than 2 MiB, so that its rows check the files the
// open opens, and the 2 MiB only for decimal.
func TestOpeningAColumnReadsLittle(t *testing.T) {
	for _, s := range surfaceSchemes() {
		for _, crash := range []bool{false, true} {
			name := s.name + "/after Close"
			if crash {
				name = s.name + "/after a crash"
			}
			t.Run(name, func(t *testing.T) {
				f := makeOpenFixture(t, s, crash)
				var chunkBytes int64
				for _, cf := range chunkFiles(t, filepath.Join(f.dir, "b", "c")) {
					chunkBytes += cf.size
				}
				for _, open := range []struct {
					name string
					open func(db *tsdb.DB) (*tsdb.Column, error)
					goOn bool
				}{
					{name: "LookupColumn", goOn: !testing.Short(), open: func(db *tsdb.DB) (*tsdb.Column, error) {
						c, _, err := db.LookupColumn("b", "c")
						return c, err
					}},
					{name: "Column", goOn: true, open: func(db *tsdb.DB) (*tsdb.Column, error) { return db.Column("b", "c", f.schema) }},
				} {
					pool := ndb.NewFilePool(1024)
					dir := crashCopy(t, f.dir) // the open may cut a torn tail; each open starts from the same disk
					db := openDB(t, dir, tsdb.Options{FilePool: pool})
					var c *tsdb.Column
					d := statsDelta(pool, func() {
						var err error
						c, err = open.open(db)
						must(t, err)
					})
					t.Logf("%s: %d files opened, %d B read", open.name, d.Opens, d.ReadBytes)
					if d.Opens > 2 {
						t.Errorf("%s opened %d files, at most schema.json and the open chunk", open.name, d.Opens)
					}
					if d.ReadBytes > 2*mib {
						t.Errorf("%s read %d B, at most %d", open.name, d.ReadBytes, 2*mib)
					}

					// counter row: the open and a query of all points need every chunk
					var n int
					var sum uint64
					q := statsDelta(pool, func() { n, sum = f.scan(t, s, c, maxTS) })
					q.Opens += d.Opens
					q.ReadBytes += d.ReadBytes
					t.Logf("%s and all points: %d files opened, %d B read of %d B of chunks", open.name, q.Opens, q.ReadBytes, chunkBytes)
					if q.Opens < int64(f.colFiles) || q.ReadBytes < chunkBytes-64*kib {
						t.Errorf("counter row: the open and a query of all points opened %d files and read %d B, want at least %d and %d",
							q.Opens, q.ReadBytes, f.colFiles, chunkBytes-64*kib)
					}

					if open.goOn {
						// the column goes on from its last intact state, also over another crash
						next := int64(10_000_000) * 20
						v := value(s, 5, 0)
						must(t, s.put(c, next, v))
						must(t, c.Sync())
						g := f
						g.extra = maps.Clone(f.extra)
						g.extra[next] = v
						again := lookup(t, openDB(t, crashCopy(t, dir), tsdb.Options{}), "b", "c")
						if n2, sum2 := g.scan(t, s, again, next); n2 != n+1 || sum2 != sum {
							t.Errorf("after a point, a Sync and another crash: %d points, want the %d of before and the new one", n2, n)
						}
					}
					closeDB(t, db)
					_ = pool.Close()
				}
			})
		}
	}
}

// TestAColumnIsOpenedOnce: the same column, which 32 goroutines open at the same time, is opened once, and each of them
// gets the same *Column, also later, until the column is dropped; after DropColumn it is another one.
func TestAColumnIsOpenedOnce(t *testing.T) {
	dir := t.TempDir()
	db := openDB(t, dir, tsdb.Options{})
	putSeries(t, column(t, db, "b", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal}), 0, 10_000)
	closeDB(t, db)

	pool := ndb.NewFilePool(1024)
	t.Cleanup(func() { _ = pool.Close() })
	db = openDB(t, dir, tsdb.Options{FilePool: pool})
	const n = 32
	got := make([]*tsdb.Column, n)
	errs := make([]error, n)
	start := make(chan struct{})
	d := statsDelta(pool, func() {
		var wg sync.WaitGroup
		for i := range n {
			wg.Go(func() {
				<-start
				if i%2 == 0 {
					got[i], _, errs[i] = db.LookupColumn("b", "c")
				} else {
					got[i], errs[i] = db.Column("b", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				}
			})
		}
		close(start)
		wg.Wait()
	})
	for i := range n {
		must(t, errs[i])
		if got[i] != got[0] {
			t.Fatalf("goroutine %d got another *Column", i)
		}
	}
	if d.Opens > 2 {
		t.Errorf("%d opens of the column opened %d files, at most schema.json and the open chunk once", n, d.Opens)
	}
	if again := lookup(t, db, "b", "c"); again != got[0] {
		t.Errorf("a later LookupColumn returns another *Column")
	}
	must(t, db.DropColumn("b", "c"))
	if fresh := column(t, db, "b", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal}); fresh == got[0] {
		t.Errorf("counter row: after DropColumn the same *Column")
	}
}

// TestEveryFileAccessGoesThroughThePool: tsdb writes and reads each file of a column through the FilePool of ndb: what a
// new column with points and a Sync leaves on disk, schema.json, the dictionary and the chunk among them, was written
// through the pool, and the open and a query of all points open each file that is not empty through the pool, the head
// with corrections included, and read at least all of its bytes. A pool of 1024 handles evicts none of them, so each is
// opened once. Renaming, cutting and removing a file are not visible in the counters and are not checked here.
func TestEveryFileAccessGoesThroughThePool(t *testing.T) {
	for _, s := range surfaceSchemes() {
		for _, head := range []bool{false, true} {
			name := s.name + "/without corrections"
			if head {
				name = s.name + "/with corrections in the head after a crash"
			}
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				pool := ndb.NewFilePool(1024)
				t.Cleanup(func() { _ = pool.Close() })
				opts := tsdb.Options{FilePool: pool, BlockPoints: 64}
				db := openDB(t, dir, opts)
				var c *tsdb.Column
				w := statsDelta(pool, func() {
					c = column(t, db, "b", "c", s.schema)
					for i := range 300 {
						must(t, s.put(c, int64(i)*10, value(s, i, 0)))
					}
					must(t, c.Sync())
				})
				onDisk := dirBytes(readSnapshot(t, filepath.Join(dir, "b", "c")))
				t.Logf("written through the pool %d B, the column has %d B", w.WrittenBytes, onDisk)
				if w.WrittenBytes < onDisk {
					t.Errorf("%d B written through the pool, the files of the column have %d B", w.WrittenBytes, onDisk)
				}
				from := dir
				if head {
					for i := range 20 {
						must(t, s.put(c, int64(i)*10, value(s, i, 1)))
					}
					from = crashCopy(t, dir)
				}
				closeDB(t, db)

				files := readSnapshot(t, filepath.Join(from, "b", "c"))
				if head && len(files["head.wal"]) == 0 {
					t.Fatalf("premise: the crash left no corrections in the head: %v", len(files["head.wal"]))
				}
				pool2 := ndb.NewFilePool(1024)
				t.Cleanup(func() { _ = pool2.Close() })
				db2 := openDB(t, from, tsdb.Options{FilePool: pool2, BlockPoints: 64})
				r := statsDelta(pool2, func() { readMap(t, s, lookup(t, db2, "b", "c")) })
				var nonEmpty []string
				for name, b := range files {
					if len(b) > 0 {
						nonEmpty = append(nonEmpty, name)
					}
				}
				slices.Sort(nonEmpty)
				t.Logf("read through the pool: %d files opened and %d B read; the column has %d B in %v", r.Opens, r.ReadBytes, dirBytes(files), nonEmpty)
				if r.Opens < int64(len(nonEmpty)) {
					t.Errorf("the open and a query of all points opened %d files through the pool, the column has %d files that are not empty: %v",
						r.Opens, len(nonEmpty), nonEmpty)
				}
				if r.ReadBytes < dirBytes(files) {
					t.Errorf("the open and a query of all points read %d B through the pool, the files of the column have %d B", r.ReadBytes, dirBytes(files))
				}
			})
		}
	}
}
