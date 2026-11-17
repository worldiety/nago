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
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The helpers of the tests in package tsdb_test, which use the exported surface of tsdb and ndb only.

const (
	minTS = math.MinInt64
	maxTS = math.MaxInt64
)

// surfaceScheme writes and reads the values of one scheme as strings, so that the tests compare decimal and enum alike.
type surfaceScheme struct {
	name   string
	schema tsdb.Schema
	put    func(c *tsdb.Column, ts int64, v string) error
	read   func(t testing.TB, c *tsdb.Column, min, max int64) ([]int64, []string)
}

func decimalScheme() surfaceScheme {
	return surfaceScheme{
		name:   "decimal",
		schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal},
		put: func(c *tsdb.Column, ts int64, v string) error {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
			return c.PutI64(ts, n)
		},
		read: func(t testing.TB, c *tsdb.Column, min, max int64) ([]int64, []string) {
			t.Helper()
			var ts []int64
			var vs []string
			if err := c.ScanI64(min, max, func(bt, bv []int64) bool {
				ts = append(ts, bt...)
				for _, v := range bv {
					vs = append(vs, strconv.FormatInt(v, 10))
				}
				return true
			}); err != nil {
				t.Fatalf("scan: %v", err)
			}
			return ts, vs
		},
	}
}

func enumScheme() surfaceScheme {
	return surfaceScheme{
		name:   "enum",
		schema: tsdb.Schema{Scheme: tsdb.SchemeEnum},
		put:    func(c *tsdb.Column, ts int64, v string) error { return c.PutString(ts, v) },
		read: func(t testing.TB, c *tsdb.Column, min, max int64) ([]int64, []string) {
			t.Helper()
			var ts []int64
			var vs []string
			if err := c.ScanString(min, max, func(bt []int64, bv []string) bool {
				ts = append(ts, bt...)
				vs = append(vs, bv...)
				return true
			}); err != nil {
				t.Fatalf("scan: %v", err)
			}
			return ts, vs
		},
	}
}

func surfaceSchemes() []surfaceScheme { return []surfaceScheme{decimalScheme(), enumScheme()} }

// value is the value of point i in a scheme: a number for decimal, one of a few names for enum. gen > 0 gives the
// value of a correction, for enum a name that is not in the dictionary yet.
func value(s surfaceScheme, i, gen int) string {
	if s.schema.Scheme == tsdb.SchemeDecimal {
		return strconv.Itoa(gen*1_000_000 + i*7%1000)
	}
	if gen == 0 {
		return "state" + strconv.Itoa(i%3)
	}
	return fmt.Sprintf("fix%d.%d", gen, i)
}

// readMap reads every point of the column as a map and fails if a time comes more than once or out of order.
func readMap(t testing.TB, s surfaceScheme, c *tsdb.Column) map[int64]string {
	t.Helper()
	ts, vs := s.read(t, c, minTS, maxTS)
	m := make(map[int64]string, len(ts))
	for i, x := range ts {
		if i > 0 && x <= ts[i-1] {
			t.Fatalf("the time %d follows %d: a point read twice or out of order", x, ts[i-1])
		}
		m[x] = vs[i]
	}
	return m
}

// quietOptions are blocks of blockPoints without compaction in the background (a ratio of tombstones above 1, a head of
// 1 GiB), for the tests that count blocks or copy the directory of a running database as the disk after a crash: a
// compaction may seal the open block at any time, and a copy taken while it renames or removes files is no state a crash
// can leave.
func quietOptions(blockPoints int) tsdb.Options {
	return tsdb.Options{BlockPoints: blockPoints, CompactTombstoneRatio: 2, FlushBytes: 1 << 30}
}

func openDB(t testing.TB, dir string, opts tsdb.Options) *tsdb.DB {
	t.Helper()
	db, err := tsdb.Open(dir, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func column(t testing.TB, db *tsdb.DB, bucket, name string, schema tsdb.Schema) *tsdb.Column {
	t.Helper()
	c, err := db.Column(bucket, name, schema)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func lookup(t testing.TB, db *tsdb.DB, bucket, name string) *tsdb.Column {
	t.Helper()
	c, ok, err := db.LookupColumn(bucket, name)
	if err != nil || !ok {
		t.Fatalf("lookup %s/%s: ok=%v err=%v", bucket, name, ok, err)
	}
	return c
}

func closeDB(t testing.TB, db *tsdb.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// crashCopy is the disk after a crash of the process: a copy of the directory, taken without Close, whose database has
// never written the memory of the running one.
func crashCopy(t testing.TB, dir string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "crashed")
	must(t, os.CopyFS(dst, os.DirFS(dir)))
	return dst
}

// snapshot holds the files of a directory, by name.
type snapshot map[string][]byte

func readSnapshot(t testing.TB, dir string) snapshot {
	t.Helper()
	entries, err := os.ReadDir(dir)
	must(t, err)
	s := snapshot{}
	for _, e := range entries {
		if e.IsDir() {
			t.Fatalf("a directory %s in the column %s", e.Name(), dir)
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		must(t, err)
		s[e.Name()] = b
	}
	return s
}

func dirBytes(s snapshot) int64 {
	var n int64
	for _, b := range s {
		n += int64(len(b))
	}
	return n
}

var (
	chunkName = regexp.MustCompile(`^[0-9]+_[0-9]+\.tsb$`)
	openName  = regexp.MustCompile(`^[0-9]+_\.tsb$`)
)

// columnFiles checks that a column holds no more files than today (d9eaf8db2): schema.json, head.wal, for enum the
// dictionary, the finalized chunks and at most one open chunk; anything else is a rest that an open or a compaction
// left behind. It returns the number of finalized and of open chunks.
func columnFiles(t testing.TB, dir string, enum bool) (final, open int) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	must(t, err)
	for _, e := range entries {
		switch name := e.Name(); {
		case name == "schema.json" || name == "head.wal" || enum && name == "enum.dict":
		case chunkName.MatchString(name):
			final++
		case openName.MatchString(name):
			open++
		default:
			t.Errorf("the column %s holds %q, which is no file of a column", dir, name)
		}
	}
	if open > 1 {
		t.Errorf("the column %s holds %d open chunks", dir, open)
	}
	return final, open
}

// liveHeap is the heap in use after a full collection.
func liveHeap() int64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.HeapAlloc)
}

// clock is the time of Options.Now, which only the test moves.
type clock struct{ ns atomic.Int64 }

func newClock() *clock {
	c := &clock{}
	c.ns.Store(time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC).UnixNano())
	return c
}

func (c *clock) now() time.Time            { return time.Unix(0, c.ns.Load()) }
func (c *clock) advance(d time.Duration)   { c.ns.Add(int64(d)) }
func (c *clock) since(t0 time.Time) string { return c.now().Sub(t0).String() }

func sortedKeys(m map[int64]string) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
