// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The acceptance tests of "keine Spalte wird größer" (wokoda N22, L-N22-24, L-N22-25): the bytes of the files per point
// of each kind of data of M22d are at most those of nago f4aca80c5 (testdata/f4aca80c5/sizes.json), and power with
// 0.1 % NA takes at most 1.55 bytes per point.

// The numbers of points of a series: in -short and in full size (without -short or with N22_FULL=1).
const (
	n22SizeShort = 1 << 16
	n22SizeFull  = 1 << 20
)

// n22Size is the size of a column after Close, as the split policy sees its chunk: tsdb.ChunkStats.SizeBytes and
// Points.
type n22Size struct {
	SizeBytes int64 `json:"sizeBytes"`
	Points    int64 `json:"points"`
}

func (s n22Size) perPoint() float64 { return float64(s.SizeBytes) / float64(s.Points) }

func n22SizeKey(kind string, compress bool, n int) string {
	return fmt.Sprintf("%s/compress=%v/%d", kind, compress, n)
}

// n22MeasureSize writes n points of the kind in the 1-s grid into a new column, closes the database and returns the
// ChunkStats of its chunk, which the split policy sees at the first point of the next open. With holes, every block
// gets one hole in its time axis: a counter row, whose time axis takes more bytes.
func n22MeasureSize(t testing.TB, k n22Kind, compress bool, n int, holes bool) n22Size {
	t.Helper()
	dir := t.TempDir()
	db := openDB(t, dir, tsdb.Options{Compress: compress})
	c := column(t, db, "n22", k.name, k.schema())
	ts := n22Base
	vals := n22Values(k.name, n)
	for i, v := range vals {
		p := n22Point{ts: ts, v: v}
		if k.enum {
			p = n22Point{ts: ts, str: n22StateName(v)}
		}
		must(t, n22Put(c, p))
		ts += 1000
		if holes && i%4096 == 100 {
			ts += 1000
		}
	}
	closeDB(t, db)

	var seen []tsdb.ChunkStats
	db = openDB(t, dir, tsdb.Options{Compress: compress, Split: func(st tsdb.ChunkStats) bool {
		seen = append(seen, st)
		return false
	}})
	c = lookup(t, db, "n22", k.name)
	if st := c.Stats(); st.Chunks != 1 {
		t.Fatalf("premise: %d chunks, the measurement takes one", st.Chunks)
	}
	if k.enum {
		must(t, c.PutString(ts, "ok"))
	} else {
		must(t, c.PutI64(ts, 1))
	}
	if len(seen) == 0 || seen[0].Points != int64(n) {
		t.Fatalf("premise: the split policy saw %+v, want the chunk of %d points", seen, n)
	}
	closeDB(t, db)
	return n22Size{SizeBytes: seen[0].SizeBytes, Points: seen[0].Points}
}

// n22SizeFindings compares the measured sizes with the reference and names each larger one.
func n22SizeFindings(got, ref map[string]n22Size) []string {
	var out []string
	for key, g := range got {
		r, ok := ref[key]
		if !ok {
			out = append(out, key+": no reference")
			continue
		}
		if g.perPoint() > r.perPoint() {
			out = append(out, fmt.Sprintf("%s: %.4f bytes per point (%d bytes), f4aca80c5 %.4f (%d bytes)", key, g.perPoint(), g.SizeBytes, r.perPoint(), r.SizeBytes))
		}
	}
	return out
}

// TestN22SizeNotLarger: no column grows (L-N22-24 "je Block nimmt nago das kleinere von neuem und heutigem Verfahren,
// sodass keine Spalte größer wird als mit f4aca80c5"; L-N22-25 "jede Datenart von M22d höchstens so viel wie
// f4aca80c5", power and power with 0.1 % NA at most 1.55 bytes per point), each kind with and without compression, in
// -short with 2^16 points, in full size with 2^20. Counter rows: the instrument sees a larger column, the same data
// without compression with a hole in every block, whose time axis takes a byte per point more, and the comparison
// fails against a reference, which is a byte smaller than f4aca80c5.
func TestN22SizeNotLarger(t *testing.T) {
	n := n22SizeShort
	if n22Full() {
		n = n22SizeFull
	}
	var ref map[string]n22Size
	n22ReadJSON(t, filepath.Join(n22FixtureRoot, "sizes.json"), &ref)
	got := map[string]n22Size{}
	holed := map[string]n22Size{}
	for _, k := range n22Kinds {
		for _, compress := range []bool{true, false} {
			key := n22SizeKey(k.name, compress, n)
			got[key] = n22MeasureSize(t, k, compress, n, false)
			t.Logf("%s: %.4f bytes per point, f4aca80c5 %.4f", key, got[key].perPoint(), ref[key].perPoint())
			if !compress {
				// with compression s2 may pack a time axis with holes smaller than one without (status, power_na1)
				holed[key] = n22MeasureSize(t, k, compress, n, true)
			}
		}
	}
	for _, f := range n22SizeFindings(got, ref) {
		t.Errorf("larger than f4aca80c5: %s", f)
	}
	for _, kind := range []string{"power", "power_na01"} {
		if s := got[n22SizeKey(kind, true, n)]; s.perPoint() > 1.55 {
			t.Errorf("%s with compression: %.4f bytes per point, want at most 1.55", kind, s.perPoint())
		}
	}
	if f := n22SizeFindings(holed, got); len(f) != len(holed) {
		t.Errorf("counter row: the instrument sees %d of %d columns with holes as larger: %v", len(f), len(holed), f)
	}
	smaller := map[string]n22Size{}
	for key, r := range ref {
		smaller[key] = n22Size{SizeBytes: r.SizeBytes - 1, Points: r.Points}
	}
	refHere := map[string]n22Size{}
	for key := range got {
		refHere[key] = ref[key]
	}
	if f := n22SizeFindings(refHere, smaller); len(f) != len(refHere) {
		t.Errorf("counter row: against a reference one byte smaller, %d of %d columns of f4aca80c5 are larger: %v", len(f), len(refHere), f)
	}
}
