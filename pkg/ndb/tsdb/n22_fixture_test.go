// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The acceptance tests of compatibility of the faster format (wokoda N22, L-N22-22 and L-N22-24) over the fixture
// testdata/f4aca80c5, which nago f4aca80c5 wrote before it: every point reads bit for bit as with f4aca80c5, old chunks
// are continued and stay as they are, and after a compaction and Close everything reads the same. The same columns,
// written by the code under test, read the same.

const n22FixtureRoot = "testdata/f4aca80c5"

// n22Put writes a point into c by its scheme.
func n22Put(c *tsdb.Column, p n22Point) error {
	if c.Schema().Scheme == tsdb.SchemeDecimal {
		return c.PutI64(p.ts, p.v)
	}
	return c.PutString(p.ts, p.str)
}

// n22WriteColumn writes the lives of col into a database at dir and returns the directory, which holds the column as
// its last life left it: dir after a Close, a copy taken before the Close after a crash. No compaction of the head runs
// in the background, so that a crash copy holds what the life wrote and nothing of a compaction in its middle.
func n22WriteColumn(t testing.TB, dir string, col n22FixCol) string {
	t.Helper()
	for _, l := range col.lives {
		db, err := tsdb.Open(dir, tsdb.Options{Compress: col.compress, CompactTombstoneRatio: 2, FlushBytes: 1 << 40})
		must(t, err)
		c := column(t, db, "n22", col.name, col.schema)
		syncAfter := map[int]bool{}
		for _, i := range l.syncAfter {
			syncAfter[i] = true
		}
		for i, p := range l.points {
			must(t, n22Put(c, p))
			if syncAfter[i] {
				must(t, c.Sync())
			}
		}
		for _, p := range l.fix {
			must(t, n22Put(c, p))
		}
		for _, ts := range l.del {
			must(t, c.Delete(ts))
		}
		if l.flush {
			must(t, c.Flush())
		}
		if l.crash {
			crashed := crashCopy(t, dir)
			closeDB(t, db)
			return crashed
		}
		closeDB(t, db)
	}
	return dir
}

// TestN22WriteFixture writes testdata/f4aca80c5 with the code of the stand it runs on: the columns of
// n22FixtureCols, what this stand reads from them (expect.json) and the bytes per point of the kinds of M22d
// (sizes.json). It ran on nago f4aca80c5, before the faster format:
//
//	N22_WRITE_FIXTURE=1 go test -run '^TestN22WriteFixture$' ./pkg/ndb/tsdb/
//
// Without N22_WRITE_FIXTURE=1 it does nothing; on a later stand it would write that stand's files instead.
func TestN22WriteFixture(t *testing.T) {
	if os.Getenv("N22_WRITE_FIXTURE") != "1" {
		t.Skip("writes testdata/f4aca80c5 only with N22_WRITE_FIXTURE=1")
	}
	must(t, os.RemoveAll(n22FixtureRoot))
	cols := n22FixtureCols()
	for _, col := range cols {
		dir := n22WriteColumn(t, t.TempDir(), col)
		must(t, os.CopyFS(filepath.Join(n22FixtureRoot, "db", "n22", col.name), os.DirFS(filepath.Join(dir, "n22", col.name))))
	}
	// what this stand reads from the fixture, which must be the model; opened from a copy, so that the open does not
	// change the fixture
	db := openDB(t, n22FixtureCopy(t), tsdb.Options{})
	reads := map[string]n22Read{}
	for _, col := range cols {
		r := n22ReadColumn(t, lookup(t, db, "n22", col.name))
		if want := n22Expect(col.schema, n22Model(col)); r != want {
			t.Errorf("%s: this stand reads %+v from the fixture, the model is %+v", col.name, r, want)
		}
		reads[col.name] = r
	}
	n22WriteJSON(t, filepath.Join(n22FixtureRoot, "expect.json"), reads)
	sizes := map[string]n22Size{}
	for _, n := range []int{n22SizeShort, n22SizeFull} {
		for _, k := range n22Kinds {
			for _, compress := range []bool{true, false} {
				sizes[n22SizeKey(k.name, compress, n)] = n22MeasureSize(t, k, compress, n, false)
			}
		}
	}
	n22WriteJSON(t, filepath.Join(n22FixtureRoot, "sizes.json"), sizes)
}

func n22WriteJSON(t testing.TB, path string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	must(t, err)
	must(t, os.WriteFile(path, append(b, '\n'), 0o644))
}

func n22ReadJSON(t testing.TB, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(path)
	must(t, err)
	must(t, json.Unmarshal(b, v))
}

// n22FixtureCopy copies the database of the fixture to a new directory and returns it.
func n22FixtureCopy(t testing.TB) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "db")
	must(t, os.CopyFS(dir, os.DirFS(filepath.Join(n22FixtureRoot, "db"))))
	return dir
}

// n22FixtureReads returns what f4aca80c5 read from each column of the fixture.
func n22FixtureReads(t testing.TB) map[string]n22Read {
	t.Helper()
	var reads map[string]n22Read
	n22ReadJSON(t, filepath.Join(n22FixtureRoot, "expect.json"), &reads)
	if len(reads) != len(n22FixtureCols()) {
		t.Fatalf("premise: expect.json holds %d columns, the fixture %d", len(reads), len(n22FixtureCols()))
	}
	return reads
}

// TestN22FixtureReads: every point of the fixture of f4aca80c5 reads bit for bit as with f4aca80c5 with ScanI64,
// ScanF64, IterF64, ScanString and IterString (L-N22-24 "Jeder Punkt liest sich bitgleich wie vorher", L-N22-22), and as
// the model of the generator says: the scaled value, NaN of math.NaN() for NA and float64(v)/10^n else.
func TestN22FixtureReads(t *testing.T) {
	reads := n22FixtureReads(t)
	db := openDB(t, n22FixtureCopy(t), tsdb.Options{})
	for _, col := range n22FixtureCols() {
		t.Run(col.name, func(t *testing.T) {
			got := n22ReadColumn(t, lookup(t, db, "n22", col.name))
			if got != reads[col.name] {
				t.Errorf("reads %+v, f4aca80c5 read %+v", got, reads[col.name])
			}
			if want := n22Expect(col.schema, n22Model(col)); got != want {
				t.Errorf("reads %+v, the model is %+v", got, want)
			}
		})
	}
}

// n22ChunkFrames returns where the blocks of a chunk file begin and the offsets of its frames, by the format of the
// README and of the block and record headers: "NAGO" "NTSB", version 1 (blocks at 9) or 2 (blocks at 85), then frames
// of an 8-byte marker, a length of 4 bytes, the region of that length and a CRC of 4 bytes over it. kinds holds 'b'
// for a block and 'r' for a continuation record. The walk ends at the first byte, which is no frame.
func n22ChunkFrames(t testing.TB, b []byte) (dataOff int, offs []int, kinds []byte) {
	t.Helper()
	if len(b) < 9 || string(b[0:4]) != "NAGO" || string(b[4:8]) != "NTSB" {
		t.Fatalf("premise: no chunk file")
	}
	switch b[8] {
	case 1:
		dataOff = 9
	case 2:
		dataOff = 85
	default:
		t.Fatalf("premise: a chunk of version %d, the test knows 1 and 2", b[8])
	}
	for at := dataOff; at+12 <= len(b); {
		var kind byte
		switch binary.LittleEndian.Uint64(b[at:]) {
		case n22BlockMarker:
			kind = 'b'
		case n22RecordMarker:
			kind = 'r'
		default:
			return dataOff, offs, kinds
		}
		end := at + 12 + int(binary.LittleEndian.Uint32(b[at+8:])) + 4
		if end > len(b) {
			return dataOff, offs, kinds
		}
		offs, kinds = append(offs, at), append(kinds, kind)
		at = end
	}
	return dataOff, offs, kinds
}

// The markers of a block and of a continuation record (block.go, chunk.go of f4aca80c5).
const (
	n22BlockMarker  uint64 = 0xDEAD54534442BEEF
	n22RecordMarker uint64 = 0xDEAD54534443BEEF
)

// n22BlocksEnd is where the blocks of a chunk file end: the first record or the end of its frames.
func n22BlocksEnd(t testing.TB, b []byte) (dataOff, end int) {
	t.Helper()
	dataOff, offs, kinds := n22ChunkFrames(t, b)
	end = dataOff
	for i, at := range offs {
		if kinds[i] != 'b' {
			break
		}
		end = at + 12 + int(binary.LittleEndian.Uint32(b[at+8:])) + 4
	}
	return dataOff, end
}

// n22MorePoints are count points of the scheme behind after, one a second.
func n22MorePoints(schema tsdb.Schema, after int64, count, seed int) []n22Point {
	vals := n22Values("power", count+seed)[seed:]
	p := make([]n22Point, count)
	for i := range p {
		p[i] = n22Point{ts: after + int64(i+1)*1000, v: vals[i]}
		if schema.Scheme != tsdb.SchemeDecimal {
			p[i] = n22Point{ts: p[i].ts, str: fmt.Sprintf("neu-%d", vals[i]%7)}
		}
	}
	return p
}

// n22Newest is the newest time of a model.
func n22Newest(pts []n22Point) int64 { return pts[len(pts)-1].ts }

// TestN22FixtureContinues: the chunks of f4aca80c5 are continued and keep their bytes (L-N22-24 "Alte Blöcke und
// Abschnitte bleiben lesbar und werden fortgesetzt ... Alte Daten werden nicht umgeschrieben"). Each column gets 1 000
// points behind its newest and a Sync in a life, which a crash ends, and 3 000 more in a life, which Close ends, which
// fills the open block of column resume, whose 500 points are in continuation records of the fixture; then every point
// reads as the model says, every finalized chunk of the fixture is the same byte for byte, and the open chunk of the
// fixture keeps its name and its blocks at their place. Column wal is read only: Close folds its head, which a
// compaction may rewrite.
func TestN22FixtureContinues(t *testing.T) {
	cols := n22FixtureCols()
	dir := n22FixtureCopy(t)
	before := map[string]snapshot{}
	for _, col := range cols {
		before[col.name] = readSnapshot(t, filepath.Join(dir, "n22", col.name))
	}
	more := map[string][]n22Point{}

	db := openDB(t, dir, tsdb.Options{Compress: true})
	for i, col := range cols {
		c := lookup(t, db, "n22", col.name)
		pts := n22MorePoints(col.schema, n22Newest(n22Model(col)), 4000, i)
		more[col.name] = pts
		for _, p := range pts[:1000] {
			must(t, n22Put(c, p))
		}
		must(t, c.Sync())
	}
	crashed := crashCopy(t, dir)
	closeDB(t, db)

	db = openDB(t, crashed, tsdb.Options{Compress: true})
	for _, col := range cols {
		c := lookup(t, db, "n22", col.name)
		for _, p := range more[col.name][1000:] {
			must(t, n22Put(c, p))
		}
	}
	closeDB(t, db)

	db = openDB(t, crashed, tsdb.Options{})
	for _, col := range cols {
		t.Run(col.name, func(t *testing.T) {
			c := lookup(t, db, "n22", col.name)
			want := n22Expect(col.schema, append(n22Model(col), more[col.name]...))
			if got := n22ReadColumn(t, c); got != want {
				t.Errorf("reads %+v, the model is %+v", got, want)
			}
			if col.name == "wal" {
				return
			}
			after := readSnapshot(t, filepath.Join(crashed, "n22", col.name))
			for name, b := range before[col.name] {
				if !strings.HasSuffix(name, ".tsb") {
					continue
				}
				now, ok := after[name]
				if !ok {
					t.Errorf("the chunk %s of the fixture is gone", name)
					continue
				}
				if !strings.HasSuffix(name, "_.tsb") {
					if !bytes.Equal(now, b) {
						t.Errorf("the finalized chunk %s of the fixture changed", name)
					}
					continue
				}
				dataOff, end := n22BlocksEnd(t, b)
				if end == dataOff {
					t.Fatalf("premise: the open chunk %s of the fixture has no block", name)
				}
				if len(now) < end || !bytes.Equal(now[dataOff:end], b[dataOff:end]) {
					t.Errorf("the blocks of the open chunk %s of the fixture changed: bytes %d to %d", name, dataOff, end)
				}
			}
		})
	}
}

// n22ForceCompaction overwrites the first and the last point of a column with its own value, so that a compaction
// rewrites all its chunks.
func n22ForceCompaction(t testing.TB, c *tsdb.Column, model []n22Point) {
	t.Helper()
	must(t, n22Put(c, model[0]))
	must(t, n22Put(c, model[len(model)-1]))
}

// TestN22FixtureCompacts: after a compaction, which rewrites every chunk of a column of the fixture (L-N22-24 "neue
// Blöcke entstehen beim Versiegeln und bei einer Verdichtung"), and Close, every point reads as with f4aca80c5. The
// compaction comes from an overwrite of the first and the last point with its own value, then Flush for every second
// column and Close for the others; premise: a chunk of the column changed.
func TestN22FixtureCompacts(t *testing.T) {
	reads := n22FixtureReads(t)
	cols := n22FixtureCols()
	dir := n22FixtureCopy(t)
	before := map[string]snapshot{}
	db := openDB(t, dir, tsdb.Options{Compress: true})
	for i, col := range cols {
		before[col.name] = readSnapshot(t, filepath.Join(dir, "n22", col.name))
		c := lookup(t, db, "n22", col.name)
		n22ForceCompaction(t, c, n22Model(col))
		if i%2 == 0 {
			must(t, c.Flush())
		}
	}
	closeDB(t, db)
	db = openDB(t, dir, tsdb.Options{})
	for _, col := range cols {
		t.Run(col.name, func(t *testing.T) {
			if got := n22ReadColumn(t, lookup(t, db, "n22", col.name)); got != reads[col.name] {
				t.Errorf("reads %+v after the compaction, f4aca80c5 read %+v", got, reads[col.name])
			}
			if !n22ChunksChanged(before[col.name], readSnapshot(t, filepath.Join(dir, "n22", col.name))) {
				t.Errorf("premise: no chunk changed, the compaction did not run")
			}
		})
	}
}

// n22ChunksChanged reports whether a chunk file of a is missing in b or differs.
func n22ChunksChanged(a, b snapshot) bool {
	for name, x := range a {
		if strings.HasSuffix(name, ".tsb") && !bytes.Equal(x, b[name]) {
			return true
		}
	}
	return false
}

// TestN22RoundTrip: the columns of the fixture, written by the code under test as f4aca80c5 wrote them, read bit for
// bit as with f4aca80c5 (expect.json) after their last life, a Close or a crash, and after a compaction of all their
// chunks and Close. This covers the new blocks of every kind of M22d, of gaps, constant blocks, the bounds of int64, NA,
// block sizes of 127, 128, 129, 4095, 4096 and 4097, enum and text.
func TestN22RoundTrip(t *testing.T) {
	reads := n22FixtureReads(t)
	for _, col := range n22FixtureCols() {
		t.Run(col.name, func(t *testing.T) {
			dir := n22WriteColumn(t, t.TempDir(), col)
			db := openDB(t, dir, tsdb.Options{Compress: col.compress})
			c := lookup(t, db, "n22", col.name)
			if got := n22ReadColumn(t, c); got != reads[col.name] {
				t.Errorf("reads %+v, f4aca80c5 read %+v", got, reads[col.name])
			}
			n22ForceCompaction(t, c, n22Model(col))
			must(t, c.Flush())
			closeDB(t, db)
			db = openDB(t, dir, tsdb.Options{})
			if got := n22ReadColumn(t, lookup(t, db, "n22", col.name)); got != reads[col.name] {
				t.Errorf("reads %+v after a compaction, f4aca80c5 read %+v", got, reads[col.name])
			}
		})
	}
}

// TestN22ScanF64: ScanF64 and IterF64 yield math.NaN() for NA and float64(v)/10^n else, bit for bit (L-N22-22), for
// every number of decimals 0 to 18, from the open block in memory, from a sealed block after Close and through the head,
// which a correction fills. The row v=3 with 1 decimal is one, in which a multiplication with 1/10^n would differ
// (premise checked), as are 13 to 36 % of the values of M22d.
func TestN22ScanF64(t *testing.T) {
	vals := []int64{0, 1, -1, 3, -3, 7, 10, 123456789, -987654321, 1<<53 + 1, -(1 << 53) - 1, math.MaxInt64, math.MinInt64 + 1,
		tsdb.NA, 999_999_999_999_999_999, 33, 1}
	if three := float64(3); three/10 == three*(1/10.0) {
		t.Fatalf("premise: 3/10 and 3*(1/10) agree, the row v=3 does not tell division from multiplication")
	}
	for dec := uint8(0); dec <= 18; dec++ {
		t.Run(fmt.Sprintf("decimals=%d", dec), func(t *testing.T) {
			dir := t.TempDir()
			db := openDB(t, dir, tsdb.Options{Compress: true})
			c := column(t, db, "n22", "f64", tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: dec})
			var pts []n22Point
			for i, v := range vals {
				pts = append(pts, n22Point{ts: n22Base + int64(i)*1000, v: v})
				must(t, c.PutI64(pts[i].ts, v))
			}
			// a correction through the head: the last point becomes 3
			pts[len(pts)-1].v = 3
			must(t, c.PutI64(pts[len(pts)-1].ts, 3))
			want := n22Expect(c.Schema(), pts)
			if got := n22ReadColumn(t, c); got != want {
				t.Errorf("the open block and the head read %+v, want %+v", got, want)
			}
			p := math.Pow10(int(dec))
			n22CheckF64(t, c, pts, p)
			closeDB(t, db)
			db = openDB(t, dir, tsdb.Options{})
			c = lookup(t, db, "n22", "f64")
			if got := n22ReadColumn(t, c); got != want {
				t.Errorf("a sealed block reads %+v, want %+v", got, want)
			}
			n22CheckF64(t, c, pts, p)
		})
	}
}

// n22CheckF64 compares each value of ScanF64 and IterF64 with float64(v)/p and NaN of NA, by its bits, and names the
// first difference.
func n22CheckF64(t *testing.T, c *tsdb.Column, pts []n22Point, p float64) {
	t.Helper()
	want := func(v int64) float64 {
		if v == tsdb.NA {
			return math.NaN()
		}
		return float64(v) / p
	}
	i := 0
	must(t, c.ScanF64(minTS, maxTS, func(ts []int64, vals []float64) bool {
		for j := range ts {
			if i < len(pts) && math.Float64bits(vals[j]) != math.Float64bits(want(pts[i].v)) {
				t.Errorf("ScanF64: point %d (%d) is %v (%#x), want %v (%#x)", i, pts[i].v, vals[j], math.Float64bits(vals[j]), want(pts[i].v), math.Float64bits(want(pts[i].v)))
			}
			i++
		}
		return true
	}))
	i = 0
	for pt := range c.IterF64(minTS, maxTS) {
		if i < len(pts) && math.Float64bits(float64(pt.Y)) != math.Float64bits(want(pts[i].v)) {
			t.Errorf("IterF64: point %d (%d) is %v, want %v", i, pts[i].v, float64(pt.Y), want(pts[i].v))
		}
		i++
	}
	if i != len(pts) {
		t.Errorf("%d points, want %d", i, len(pts))
	}
}
