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
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The acceptance tests of R-DEC-TS-NAGO-DICT (wokoda N23, L-N23-5, L-N23-31): a point of an enum never reads as another
// text than the one it was written with, also after a damage or a cut of enum.dict, after a reopen, after the rewrite of
// a dictionary in the format before N23 and after a new value came; a point, whose text is lost or missing, is missing
// like a point of a damaged block, and that is reported once per place (R-DEC-TS-NAGO-DAMAGE).
//
// The tests write the values A, B, C and D (n23Value, at least 1 000 bytes each), with points in a sealed block, in a
// continuation record and D as a correction in the WAL (n23WriteDict, a crash copy), damage enum.dict with os, open
// again and read every point; then they write a new value E, read again and open once more. A row with "E first" writes
// E before the first read of an old point. The start of an entry comes from the size of enum.dict after the value was
// written first; in a dictionary of N23 its text lies behind its header of n23DictEntryHdr bytes, masked (L-N23-36), in
// an image in the format before N23 a test finds it with bytes.Index.

const n23E = int64(20000) // the time of E

// n23DictState is what a point of one of the values A to D reads.
type n23DictState string

const (
	n23As     n23DictState = "as written"
	n23Gone   n23DictState = "missing"
	n23Either n23DictState = "as written or missing"
	n23Empty  n23DictState = "the empty text"
)

// n23CheckDict compares what the column read (got) with the model of n23DictModel: each point of A to D by want, E at
// n23E with the text e, if e is not empty, and no other time. No point reads as another text than its own, E or, where
// want says so, the empty text.
func n23CheckDict(t *testing.T, when string, got map[int64]string, want map[string]n23DictState, e string) {
	t.Helper()
	model := n23DictModel()
	for ts, name := range model {
		v, present := got[ts]
		own := n23Value(name)
		var ok bool
		switch want[name] {
		case n23As:
			ok = present && v == own
		case n23Gone:
			ok = !present
		case n23Either:
			ok = !present || v == own
		case n23Empty:
			ok = present && v == ""
		default:
			t.Fatalf("no state for %s", name)
		}
		if !ok {
			t.Errorf("%s: the point %d of %s reads %q (present %v), want %s", when, ts, name, n23Short(v), present, want[name])
		}
	}
	for ts, v := range got {
		if _, old := model[ts]; old {
			continue
		}
		if ts == n23E && e != "" {
			if v != e {
				t.Errorf("%s: E at %d reads %q, want E", when, ts, n23Short(v))
			}
			continue
		}
		t.Errorf("%s: a point %d=%q, which nobody wrote", when, ts, n23Short(v))
	}
	if _, ok := got[n23E]; e != "" && !ok {
		t.Errorf("%s: E at %d is missing", when, n23E)
	}
}

// n23Short is a text for a message: the name of a value, or its first bytes.
func n23Short(s string) string {
	if n := n23NameOf(s); n != s {
		return n
	}
	if len(s) > 12 {
		return s[:12] + "…"
	}
	return s
}

// n23ReadEnum reads every point of the enum column c.
func n23ReadEnum(t *testing.T, c *tsdb.Column) map[int64]string {
	t.Helper()
	got, err := scanMap(enumScheme(), c)
	if err != nil {
		t.Fatalf("ScanString: %v", err)
	}
	return got
}

// n23Warnings are the warnings of the bucket and the column since the index from of the records of h.
func n23Warnings(h *n22cHandler, bucket, column string, from int) []n22cRecord {
	all := h.warnings(bucket, column)
	if from > len(all) {
		return nil
	}
	return all[from:]
}

// n23Offsets are the offsets of the records, which must name file.
func n23Offsets(t *testing.T, when string, recs []n22cRecord, file string) []int64 {
	t.Helper()
	var offs []int64
	for _, r := range recs {
		off, err := strconv.ParseInt(r.attrs["offset"], 10, 64)
		if filepath.Base(r.attrs["file"]) != file || err != nil {
			t.Errorf("%s: a warning %v, want file %s and a number as offset", when, r.attrs, file)
		}
		offs = append(offs, off)
	}
	return offs
}

// n23DictImage copies the image of column col of testdata/8ae1daaee to the column bucket/d of a new database and
// returns the database directory.
func n23DictImage(t *testing.T, col, bucket string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "db")
	must(t, os.CopyFS(filepath.Join(dir, bucket, "d"), os.DirFS(filepath.Join(n23FixtureRoot, "db", "n23", col))))
	return dir
}

// n23DictAt are the starts of the entries of A to D and "end", the end of D, in a dictionary: for one of the code under
// test from the size of enum.dict after each value was written first ("A" unknown, -1), for an image by the layout of
// 8ae1daaee.
func n23OldDictAt(t *testing.T, b []byte) map[string]int {
	t.Helper()
	offs, texts := n23OldDictEntries(b)
	at := map[string]int{"end": len(b)}
	for i, name := range []string{"A", "B", "C", "D"} {
		if i >= len(offs) || texts[i] != n23Value(name) {
			t.Fatalf("premise: the image holds the entries %d, the %d-th is not %s", len(offs), i, name)
		}
		at[name] = offs[i]
	}
	return at
}

func n23NewDictAt(ends map[string]int64) map[string]int {
	return map[string]int{"A": int(ends["A"]) - n23DictEntryHdr - n23DictLens["A"], "B": int(ends["A"]), "C": int(ends["B"]), "D": int(ends["C"]), "end": int(ends["D"])}
}

// n23DictCase is a row of the tables of R-DEC-TS-NAGO-DICT.
type n23DictCase struct {
	name string
	old  bool // the image dict of 8ae1daaee instead of a dictionary of the code under test
	// edit damages the dictionary b, whose entries begin at at, and returns it; nil: the counter row
	edit   func(t *testing.T, b []byte, at map[string]int) []byte
	eFirst bool // E before the first read of an old point
	noE    bool // no E: an exception of the contract, for which a row with E would hold an exchange as behaviour
	eLen   int  // the length of E; 0: 1 000
	want   map[string]n23DictState
	// reports are the warnings of the column in the first and in the second life of the DB; atLoad: the first comes
	// when the dictionary loads, before a point of a damaged entry is read
	reports [2]int
	atLoad  bool
	// offset is where the report of the first life lies; nil: not checked. again bounds the one of the second life by
	// the dictionary after the first life; nil: not checked
	offset func(at map[string]int) int64
	again  func(t *testing.T, dict []byte) (lo, hi int64)
	// keep: the entries of these values stay as they are at their place in enum.dict after E and Close ("kein Schreiben
	// überschreibt sie")
	keep []string
}

func n23All(s n23DictState) map[string]n23DictState {
	return map[string]n23DictState{"A": s, "B": s, "C": s, "D": s}
}

func n23Want(base n23DictState, others ...any) map[string]n23DictState {
	w := n23All(base)
	for i := 0; i+1 < len(others); i += 2 {
		w[others[i].(string)] = others[i+1].(n23DictState)
	}
	return w
}

// n23Flip flips the byte at i.
func n23Flip(b []byte, i int) []byte { b[i] ^= 0x5A; return b }

// n23RunDict runs a row of the tables of R-DEC-TS-NAGO-DICT.
func n23RunDict(t *testing.T, h *n22cHandler, bucket string, c n23DictCase) {
	var dir string
	var at map[string]int
	if c.old {
		dir = n23DictImage(t, "dict", bucket)
	} else {
		var ends map[string]int64
		dir, ends = n23WriteDict(t, t.TempDir(), bucket, "d", false)
		at = n23NewDictAt(ends)
	}
	path := filepath.Join(dir, bucket, "d", "enum.dict")
	b, err := os.ReadFile(path)
	must(t, err)
	if c.old {
		at = n23OldDictAt(t, b)
	} else if len(b) != at["end"] {
		t.Fatalf("premise: enum.dict has %d bytes, the end of D was at %d", len(b), at["end"])
	}
	orig := slices.Clone(b)
	if c.edit != nil {
		b = c.edit(t, b, at)
		if b == nil {
			must(t, os.Remove(path))
		} else {
			must(t, os.WriteFile(path, b, 0o644))
		}
	}
	e := ""
	if !c.noE {
		n := c.eLen
		if n == 0 {
			n = 1000
		}
		e = n23Text("E", n)
	}
	nextOf := map[string]string{"A": "B", "B": "C", "C": "D", "D": "end"}
	kept := map[string][]byte{}
	for _, name := range c.keep {
		kept[name] = slices.Clone(orig[at[name]:at[nextOf[name]]])
	}

	life := func(n int, from int) {
		recs := n23Warnings(h, bucket, "d", from)
		if len(recs) != c.reports[n] {
			t.Errorf("life %d: %d warnings of the column, want %d: %v", n+1, len(recs), c.reports[n], recs)
		}
	}
	db := openDB(t, dir, n23DictOptions())
	col := lookup(t, db, bucket, "d")
	if c.eFirst {
		must(t, col.PutString(n23E, e))
		if got := len(h.warnings(bucket, "d")); c.atLoad && got != c.reports[0] {
			t.Errorf("after the write of E, which loads the dictionary: %d warnings, want %d", got, c.reports[0])
		}
		n23CheckDict(t, "after E first", n23ReadEnum(t, col), c.want, e)
	} else {
		// a read of the point of A at 1 000 loads the dictionary
		must(t, col.ScanString(1000, 1000, func([]int64, []string) bool { return true }))
		if got := len(h.warnings(bucket, "d")); c.atLoad && got != c.reports[0] {
			t.Errorf("after the first read, which loads the dictionary: %d warnings, want %d", got, c.reports[0])
		}
		n23CheckDict(t, "the first read", n23ReadEnum(t, col), c.want, "")
		n23CheckDict(t, "the second read", n23ReadEnum(t, col), c.want, "")
		if !c.noE {
			must(t, col.PutString(n23E, e))
			n23CheckDict(t, "after E", n23ReadEnum(t, col), c.want, e)
		}
	}
	first := h.warnings(bucket, "d")
	life(0, 0)
	if c.offset != nil && len(first) > 0 {
		offs := n23Offsets(t, "life 1", first, "enum.dict")
		if want := c.offset(at); offs[0] != want {
			t.Errorf("life 1: the report has offset %d, want %d", offs[0], want)
		}
	}
	closeDB(t, db)
	afterLife1, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, name := range c.keep {
		if end := at[nextOf[name]]; len(afterLife1) < end || !bytes.Equal(afterLife1[at[name]:end], kept[name]) {
			t.Errorf("the entry of %s at %d changed in enum.dict: a write went over it", name, at[name])
		}
	}

	mark := len(h.warnings(bucket, "d"))
	db = openDB(t, dir, n23DictOptions())
	col = lookup(t, db, bucket, "d")
	n23CheckDict(t, "after reopening", n23ReadEnum(t, col), c.want, e)
	life(1, mark)
	if second := n23Warnings(h, bucket, "d", mark); c.again != nil && len(second) > 0 {
		offs := n23Offsets(t, "life 2", second, "enum.dict")
		if lo, hi := c.again(t, afterLife1); offs[0] < lo || offs[0] > hi {
			t.Errorf("life 2: the report has offset %d, want it at the place of the entry in the new dictionary, %d … %d", offs[0], lo, hi)
		}
	}
	closeDB(t, db)
}

// TestN23DictDamage is the first table of R-DEC-TS-NAGO-DICT over dictionaries, which the code under test wrote.
func TestN23DictDamage(t *testing.T) {
	h := n22cCapture(t)
	cutTail := func(t *testing.T, b []byte, at map[string]int) []byte {
		if at["end"]-at["D"] <= 500 {
			t.Fatalf("premise: the entry of D (%d bytes) is not longer than 500 bytes", at["end"]-at["D"])
		}
		return b[:len(b)-500]
	}
	rows := []n23DictCase{
		{name: "counter row: an intact dictionary", want: n23All(n23As)},
		{name: "a byte in the middle of the text of B, its length intact",
			edit: func(t *testing.T, b []byte, at map[string]int) []byte {
				return n23Flip(b, at["B"]+n23DictEntryHdr+n23DictLens["B"]/2)
			},
			want: n23Want(n23As, "B", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(at map[string]int) int64 { return int64(at["B"]) }, keep: []string{"A", "C", "D"}},
		{name: "the same, but E first", eFirst: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte {
				return n23Flip(b, at["B"]+n23DictEntryHdr+n23DictLens["B"]/2)
			},
			want: n23Want(n23As, "B", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(at map[string]int) int64 { return int64(at["B"]) }, keep: []string{"A", "C", "D"}},
		{name: "a byte in the length of B, so that B ends within the dictionary",
			edit: func(t *testing.T, b []byte, at map[string]int) []byte {
				off, size := n23DictLenField(t, b, at["B"])
				if size == 0 {
					return b
				}
				return n23Flip(b, off) // the lowest byte: B ends some bytes off its end, within C
			},
			want: n23Want(n23Either, "A", n23As, "B", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(at map[string]int) int64 { return int64(at["B"]) }},
		{name: "B and C with their headers zeroed",
			edit: func(t *testing.T, b []byte, at map[string]int) []byte {
				clear(b[at["B"]:at["D"]])
				return b
			},
			want: n23Want(n23Either, "A", n23As, "B", n23Gone, "C", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(at map[string]int) int64 { return int64(at["B"]) }},
		{name: "the same, but E first", eFirst: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte {
				clear(b[at["B"]:at["D"]])
				return b
			},
			want: n23Want(n23Either, "A", n23As, "B", n23Gone, "C", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(at map[string]int) int64 { return int64(at["B"]) }},
		{name: "the last 500 bytes cut off, in the middle of D", edit: cutTail,
			want: n23Want(n23As, "D", n23Gone), reports: [2]int{1, 1},
			offset: func(at map[string]int) int64 { return int64(at["D"]) }},
		{name: "the same, but E first", eFirst: true, edit: cutTail,
			want: n23Want(n23As, "D", n23Gone), reports: [2]int{1, 1},
			offset: func(at map[string]int) int64 { return int64(at["D"]) }},
		{name: "cut right before D, its points in a sealed block, a continuation record and the WAL", noE: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte { return b[:at["D"]] },
			want: n23Want(n23As, "D", n23Gone), reports: [2]int{1, 1},
			offset: func(at map[string]int) int64 { return int64(at["D"]) },
			again:  func(t *testing.T, d []byte) (int64, int64) { return int64(len(d)), int64(len(d)) }},
		{name: "enum.dict removed", noE: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte { return nil },
			want: n23All(n23Gone), reports: [2]int{1, 1},
			offset: func(map[string]int) int64 { return 0 },
			again:  func(*testing.T, []byte) (int64, int64) { return 0, 0 }},
	}
	for i, c := range rows {
		t.Run(c.name, func(t *testing.T) { n23RunDict(t, h, "dictdamage"+strconv.Itoa(i), c) })
	}

	t.Run("the last 500 bytes cut off, but the points of D deleted before", func(t *testing.T) {
		const bucket = "dictdamagedel"
		dir := t.TempDir()
		db := openDB(t, dir, n23DictOptions())
		col := column(t, db, bucket, "d", tsdb.Schema{Scheme: tsdb.SchemeEnum})
		var dictEnd int64
		path := filepath.Join(dir, bucket, "d", "enum.dict")
		sealed, record, wal := n23DictPoints()
		for _, p := range slices.Concat(sealed, record, []n22Point{wal}) {
			must(t, col.PutString(p.ts, p.str))
			dictEnd = n23Size(t, path)
		}
		model := n23DictModel()
		for ts, name := range model {
			if name == "D" {
				must(t, col.Delete(ts))
				delete(model, ts)
			}
		}
		closeDB(t, db) // folds the head: no file of the column holds a point of D
		b, err := os.ReadFile(path)
		must(t, err)
		if int64(len(b)) != dictEnd {
			t.Fatalf("premise: enum.dict has %d bytes, %d after D", len(b), dictEnd)
		}
		must(t, os.WriteFile(path, b[:len(b)-500], 0o644))
		e := n23Text("E", 1000)
		for n := range 2 {
			db := openDB(t, dir, n23DictOptions())
			col := lookup(t, db, bucket, "d")
			if n == 0 {
				must(t, col.PutString(n23E, e))
			}
			got := n23ReadEnum(t, col)
			want := map[int64]string{n23E: e}
			for ts, name := range model {
				want[ts] = n23Value(name)
			}
			if !mapsEqual(got, want) {
				t.Errorf("life %d: reads %v, want %v", n+1, n23Points(got), n23Points(want))
			}
			closeDB(t, db)
		}
		if w := h.warnings(bucket, "d"); len(w) != 0 {
			t.Errorf("%d warnings, want none: %v", len(w), w)
		}
	})

	t.Run("an intact dictionary with a version after N23", func(t *testing.T) {
		const bucket = "dictdamagever"
		dir, _ := n23WriteDict(t, t.TempDir(), bucket, "d", false)
		path := filepath.Join(dir, bucket, "d", "enum.dict")
		b, err := os.ReadFile(path)
		must(t, err)
		n23DictBumpVersion(t, b)
		must(t, os.WriteFile(path, b, 0o644))
		db := openDB(t, dir, n23DictOptions())
		col := lookup(t, db, bucket, "d")
		if err := col.ScanString(minTS, maxTS, func([]int64, []string) bool { return true }); err == nil {
			t.Errorf("a read of the column succeeds over a dictionary of a later version")
		}
		if err := col.PutString(n23E, n23Text("E", 1000)); err == nil {
			t.Errorf("a write of a new value succeeds over a dictionary of a later version")
		}
		if err := col.PutString(n23E+1000, n23Value("A")); err == nil {
			t.Errorf("a write of a known value succeeds over a dictionary of a later version")
		}
		_ = db.Close()
		if now, err := os.ReadFile(path); err != nil || !bytes.Equal(now, b) {
			t.Errorf("enum.dict changed after the attempts and Close (%v)", err)
		}
		if w := h.warnings(bucket, "d"); len(w) != 0 {
			t.Errorf("%d warnings, want none: %v", len(w), w)
		}
	})
}

// TestN23OldDict is the table "Wörterbuch im Format vor N23" of R-DEC-TS-NAGO-DICT over the image dict of 8ae1daaee.
func TestN23OldDict(t *testing.T) {
	h := n22cCapture(t)
	reads := n23FixtureReads(t)
	if want := n23Points(n23DictModel()); !slices.Equal(reads["dict"], want) {
		t.Fatalf("premise: 8ae1daaee read %v from the image, the model is %v", reads["dict"], want)
	}
	// newAt bounds the place of the entry of A (id 0) or B (id 1) in the rewritten dictionary by the entries around it,
	// found by their unmasked headers
	newAt := func(name string) func(t *testing.T, d []byte) (int64, int64) {
		return func(t *testing.T, d []byte) (int64, int64) {
			ids := n23DictLayout(t).entries(d)
			at := func(id uint32) int64 {
				a, ok := ids[id]
				if !ok {
					t.Fatalf("premise: the rewritten dictionary has no entry with the id %d: %v", id, ids)
				}
				return int64(a)
			}
			switch name {
			case "A":
				return 0, at(1)
			case "B":
				return at(0) + 1, at(2)
			}
			t.Fatalf("no place of %s", name)
			return 0, 0
		}
	}
	rows := []n23DictCase{
		{name: "counter row: an intact image", old: true, want: n23All(n23As)},
		{name: "an intact image, E first", old: true, eFirst: true, want: n23All(n23As)},
		{name: "B and C with their headers zeroed, which the old format reads as intact empty entries", old: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte {
				if (at["D"]-at["B"])%n23OldDictEntryHdr != 0 {
					t.Fatalf("premise: the entries of B and C (%d bytes) are no multiple of the header", at["D"]-at["B"])
				}
				clear(b[at["B"]:at["D"]])
				if offs, _ := n23OldDictEntries(b); len(offs) != 1+(at["D"]-at["B"])/n23OldDictEntryHdr+1 {
					t.Fatalf("premise: the old format reads %d entries, not A, the empty ones and D", len(offs))
				}
				return b
			},
			want: n23Want(n23Empty, "A", n23As)},
		{name: "a byte in the middle of the text of B, its length intact, then E as long as B", old: true,
			eLen: n23DictLens["B"],
			edit: func(t *testing.T, b []byte, at map[string]int) []byte {
				return n23Flip(b, at["B"]+n23OldDictEntryHdr+n23DictLens["B"]/2)
			},
			want: n23Want(n23As, "B", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(at map[string]int) int64 { return int64(at["B"]) }, again: newAt("B")},
		{name: "a byte in the CRC of A, where nago tells the formats apart, then E", old: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte { return n23Flip(b, 4) },
			want: n23Want(n23As, "A", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(map[string]int) int64 { return 0 }, again: newAt("A")},
		{name: "a byte in the middle of the text of A, then E", old: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte {
				return n23Flip(b, n23OldDictEntryHdr+n23DictLens["A"]/2)
			},
			want: n23Want(n23As, "A", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(map[string]int) int64 { return 0 }, again: newAt("A")},
		{name: "a byte in the length of B, E first", old: true, eFirst: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte { return n23Flip(b, at["B"]) },
			want: n23Want(n23Either, "A", n23As, "B", n23Gone), reports: [2]int{1, 1}, atLoad: true,
			offset: func(at map[string]int) int64 { return int64(at["B"]) }},
		{name: "the last 500 bytes cut off, in the middle of D, E first", old: true, eFirst: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte { return b[:len(b)-500] },
			want: n23Want(n23As, "D", n23Gone), reports: [2]int{1, 1}},
		{name: "cut right before D, its points in a sealed block", old: true, noE: true,
			edit: func(t *testing.T, b []byte, at map[string]int) []byte { return b[:at["D"]] },
			want: n23Want(n23As, "D", n23Gone), reports: [2]int{1, 1}},
	}
	for i, c := range rows {
		t.Run(c.name, func(t *testing.T) {
			bucket := "olddict" + strconv.Itoa(i)
			n23RunDict(t, h, bucket, c)
		})
	}

	t.Run("an intact image is rewritten into the format of N23, E in it", func(t *testing.T) {
		const bucket = "olddictformat"
		dir := n23DictImage(t, "dict", bucket)
		path := filepath.Join(dir, bucket, "d", "enum.dict")
		old, err := os.ReadFile(path)
		must(t, err)
		db := openDB(t, dir, n23DictOptions())
		col := lookup(t, db, bucket, "d")
		e := n23Text("E", 1000)
		must(t, col.PutString(n23E, e))
		closeDB(t, db)
		b, err := os.ReadFile(path)
		must(t, err)
		if bytes.Equal(b, old) || bytes.HasPrefix(b, old) {
			t.Errorf("enum.dict is the old one or continues it: it is not rewritten")
		}
		// E does not stand in an entry of the old format, and its text is masked (L-N23-36)
		if offs, texts := n23OldDictEntries(b); slices.Contains(texts, e) {
			t.Errorf("E stands in enum.dict in the format before N23 (entries at %v)", offs)
		}
		if bytes.Contains(b, []byte(e)) || bytes.Contains(b, []byte(n23Value("A"))) {
			t.Errorf("enum.dict holds the text of E or A as written: the rewritten dictionary does not mask its texts")
		}
		// counter row of the instrument: the image before the rewrite holds the text of A as written
		if !bytes.Contains(old, []byte(n23Value("A"))) {
			t.Errorf("premise: the image holds the text of A not as written")
		}
		if !n23DictIsNew(t, b) {
			t.Errorf("enum.dict is not in the format of N23 after the rewrite")
		}
		if w := h.warnings(bucket, "d"); len(w) != 0 {
			t.Errorf("%d warnings, want none: %v", len(w), w)
		}
	})

	t.Run("a dictionary in the format of N23, damaged at the place, where nago tells the formats apart", func(t *testing.T) {
		const bucket = "olddictplace"
		dir, _ := n23WriteDict(t, t.TempDir(), bucket, "d", false)
		path := filepath.Join(dir, bucket, "d", "enum.dict")
		b, err := os.ReadFile(path)
		must(t, err)
		at := n23DictFormatPlace(t, b)
		n23Flip(b, at)
		must(t, os.WriteFile(path, b, 0o644))
		db := openDB(t, dir, n23DictOptions())
		col := lookup(t, db, bucket, "d")
		n23CheckDict(t, "the first read", n23ReadEnum(t, col), n23All(n23As), "")
		e := n23Text("E", 1000)
		must(t, col.PutString(n23E, e))
		n23CheckDict(t, "after E", n23ReadEnum(t, col), n23All(n23As), e)
		closeDB(t, db)
		now, err := os.ReadFile(path)
		must(t, err)
		if !bytes.HasPrefix(now, b) {
			t.Errorf("enum.dict is rewritten like one in the format before N23 (it does not continue the damaged file)")
		}
		w := h.warnings(bucket, "d")
		if len(w) != 1 {
			t.Errorf("%d warnings, want one: %v", len(w), w)
		} else if offs := n23Offsets(t, "the report", w, "enum.dict"); offs[0] != int64(at) {
			t.Errorf("the report has offset %d, want %d", offs[0], at)
		}
	})

	t.Run("a crash during the rewrite of the intact image", func(t *testing.T) {
		const bucket = "olddictcrash"
		names := n23DictRewriteFiles(t)
		// the files of the rewrite as it writes them: from a run of the rewrite on a copy
		full := n23DictImage(t, "dict", bucket)
		colDir := filepath.Join(full, bucket, "d")
		old := readSnapshot(t, colDir)
		db := openDB(t, full, n23DictOptions())
		n23CheckDict(t, "the rewrite", n23ReadEnum(t, lookup(t, db, bucket, "d")), n23All(n23As), "")
		closeDB(t, db)
		newDict, err := os.ReadFile(filepath.Join(colDir, "enum.dict"))
		must(t, err)
		// the states: each file of the rewrite cut at every length between 0 and its full size beside the old
		// enum.dict, and the states between two of its steps, in which enum.dict is the old or the new dictionary
		var states []snapshot
		for _, name := range names {
			for n := 0; n <= len(newDict); n++ {
				if testing.Short() && n != 0 && n != len(newDict)/2 && n != len(newDict) {
					continue
				}
				s := cloneSnapshot(old)
				s[name] = newDict[:n]
				states = append(states, s)
			}
		}
		// the states between two steps, in which enum.dict is the old dictionary (above: beside each file of the
		// rewrite, also at its full length) or the new one, without a file of the rewrite
		states = append(states, cloneSnapshot(old))
		withNew := cloneSnapshot(old)
		withNew["enum.dict"] = newDict
		states = append(states, withNew)
		for k, s := range states {
			dir := filepath.Join(t.TempDir(), "db")
			cd := filepath.Join(dir, bucket, "d")
			must(t, os.MkdirAll(cd, 0o755))
			for name, b := range s {
				must(t, os.WriteFile(filepath.Join(cd, name), b, 0o644))
			}
			mark := len(h.warnings(bucket, "d"))
			db := openDB(t, dir, n23DictOptions())
			col := lookup(t, db, bucket, "d")
			e := n23Text("E", 1000)
			n23CheckDict(t, fmt.Sprintf("state %d", k), n23ReadEnum(t, col), n23All(n23As), "")
			must(t, col.PutString(n23E, e))
			n23CheckDict(t, fmt.Sprintf("state %d after E", k), n23ReadEnum(t, col), n23All(n23As), e)
			closeDB(t, db)
			after := readSnapshot(t, cd)
			for _, name := range names {
				if _, ok := after[name]; ok {
					t.Errorf("state %d: the file %s of the rewrite stays in the column", k, name)
				}
			}
			if !n23DictIsNew(t, after["enum.dict"]) {
				t.Errorf("state %d: enum.dict is not in the format of N23 after the open", k)
			}
			if w := n23Warnings(h, bucket, "d", mark); len(w) != 0 {
				t.Errorf("state %d: %d warnings, want none: %v", k, len(w), w)
			}
		}
	})

	t.Run("a dictionary rewritten by N23 and one created by N23 are not rewritten again", func(t *testing.T) {
		for _, src := range []string{"rewritten", "created"} {
			t.Run(src, func(t *testing.T) {
				bucket := "olddictsame" + src
				var dir string
				if src == "rewritten" {
					dir = n23DictImage(t, "dict", bucket)
				} else {
					dir, _ = n23WriteDict(t, t.TempDir(), bucket, "d", false)
				}
				path := filepath.Join(dir, bucket, "d", "enum.dict")
				image, err := os.ReadFile(path)
				must(t, err)
				// the first life: for the image the rewrite, for the created one a read
				db := openDB(t, dir, n23DictOptions())
				n23CheckDict(t, "the first life", n23ReadEnum(t, lookup(t, db, bucket, "d")), n23All(n23As), "")
				closeDB(t, db)
				held, err := os.Open(path)
				must(t, err)
				defer func() { _ = held.Close() }()
				heldInfo, err := held.Stat()
				must(t, err)
				before, err := os.ReadFile(path)
				must(t, err)
				if src == "rewritten" {
					// counter row: the first open of the intact image replaced or changed enum.dict
					if bytes.Equal(before, image) {
						t.Errorf("premise: the first open of the image left enum.dict as it was; it was not rewritten")
					}
				}
				db = openDB(t, dir, n23DictOptions())
				col := lookup(t, db, bucket, "d")
				n23CheckDict(t, "the second life", n23ReadEnum(t, col), n23All(n23As), "")
				now, err := os.Stat(path)
				must(t, err)
				nowBytes, err := os.ReadFile(path)
				must(t, err)
				if !os.SameFile(heldInfo, now) || !bytes.Equal(nowBytes, before) {
					t.Errorf("the second life rewrote enum.dict (same file %v, same bytes %v)", os.SameFile(heldInfo, now), bytes.Equal(nowBytes, before))
				}
				e := n23Text("E", 1000)
				must(t, col.PutString(n23E, e))
				closeDB(t, db)
				now, err = os.Stat(path)
				must(t, err)
				nowBytes, err = os.ReadFile(path)
				must(t, err)
				if !os.SameFile(heldInfo, now) || !bytes.HasPrefix(nowBytes, before) {
					t.Errorf("after E enum.dict is not the same file with its bytes of before at its start (same file %v)", os.SameFile(heldInfo, now))
				}
				if w := h.warnings(bucket, "d"); len(w) != 0 {
					t.Errorf("%d warnings, want none: %v", len(w), w)
				}
			})
		}
	})

	t.Run("the rewrite is kept from writing", func(t *testing.T) {
		const bucket = "olddictblocked"
		names := n23DictRewriteFiles(t)
		dir := n23DictImage(t, "dict", bucket)
		colDir := filepath.Join(dir, bucket, "d")
		before := readSnapshot(t, colDir)
		blocker := filepath.Join(colDir, names[0])
		must(t, os.Mkdir(blocker, 0o755))
		must(t, os.WriteFile(filepath.Join(blocker, "x"), []byte("x"), 0o644))
		db := openDB(t, dir, n23DictOptions())
		col := lookup(t, db, bucket, "d")
		if err := col.ScanString(minTS, maxTS, func([]int64, []string) bool { return true }); err == nil {
			t.Errorf("a read succeeds, although the rewrite cannot write %s", names[0])
		}
		if b, err := os.ReadFile(filepath.Join(colDir, "enum.dict")); err != nil || !bytes.Equal(b, before["enum.dict"]) {
			t.Errorf("enum.dict changed although the rewrite failed (%v)", err)
		}
		// counter row: without the obstacle the next access reads as written
		must(t, os.RemoveAll(blocker))
		n23CheckDict(t, "without the obstacle", n23ReadEnum(t, col), n23All(n23As), "")
		closeDB(t, db)
	})

	t.Run("a read, which rewrites the dictionary, changes no other file of the column", func(t *testing.T) {
		const bucket = "olddictonly"
		dir := n23DictImage(t, "dict", bucket)
		colDir := filepath.Join(dir, bucket, "d")
		db := openDB(t, dir, n23DictOptions())
		col := lookup(t, db, bucket, "d")
		before := readSnapshot(t, colDir)
		if len(before["head.wal"]) == 0 {
			t.Fatalf("premise: the image has no WAL in the format before N23")
		}
		n23CheckDict(t, "ScanString", n23ReadEnum(t, col), n23All(n23As), "")
		after := readSnapshot(t, colDir)
		for _, d := range n22SnapshotDiff(before, after) {
			if !strings.HasPrefix(d, "enum.dict ") {
				t.Errorf("the read changed another file of the column: %s", d)
			}
		}
		// counter row: enum.dict changed, it is rewritten
		if bytes.Equal(before["enum.dict"], after["enum.dict"]) {
			t.Errorf("enum.dict did not change: the read did not rewrite it")
		}
		closeDB(t, db)
		if w := h.warnings(bucket, "d"); len(w) != 0 {
			t.Errorf("%d warnings, want none: %v", len(w), w)
		}
	})
}

// TestN23OldDictTornTail: an image in the format before N23, whose last entry X carries no point yet (dicttail), cut
// at every length at which X has begun and not ended, as after a crash under N22 while X was created; then N23 opens it
// (R-DEC-TS-NAGO-DICT, "Was nur anhängt, prüft dazu der Test über abgeschnittene Dateien"). After the rewrite and after a
// reopen no report comes, the other points read as written, a new value Y reads as itself, and no rest of X is read
// as an entry. With -short a few lengths only.
func TestN23OldDictTornTail(t *testing.T) {
	h := n22cCapture(t)
	src := filepath.Join(n23FixtureRoot, "db", "n23", "dicttail", "enum.dict")
	b, err := os.ReadFile(src)
	must(t, err)
	offs, texts := n23OldDictEntries(b)
	if len(offs) != 5 || texts[4] != n23Value("X") {
		t.Fatalf("premise: the image holds %d entries, the last not X", len(offs))
	}
	from := offs[4]
	var lengths []int
	for n := from + 1; n < len(b); n++ {
		if !testing.Short() || n == from+1 || n == from+n23OldDictEntryHdr || n == (from+len(b))/2 || n == len(b)-1 {
			lengths = append(lengths, n)
		}
	}
	y := n23Text("Y", 1000)
	for _, n := range lengths {
		bucket := "dicttail" + strconv.Itoa(n)
		dir := filepath.Join(t.TempDir(), "db")
		cd := filepath.Join(dir, bucket, "d")
		must(t, os.CopyFS(cd, os.DirFS(filepath.Dir(src))))
		must(t, os.WriteFile(filepath.Join(cd, "enum.dict"), b[:n], 0o644))
		for life := range 2 {
			db := openDB(t, dir, n23DictOptions())
			col := lookup(t, db, bucket, "d")
			if life == 0 {
				n23CheckDict(t, fmt.Sprintf("cut at %d", n), n23ReadEnum(t, col), n23All(n23As), "")
				must(t, col.PutString(n23E, y))
			}
			n23CheckDict(t, fmt.Sprintf("cut at %d, life %d with Y", n, life+1), n23ReadEnum(t, col), n23All(n23As), y)
			closeDB(t, db)
		}
		if w := h.warnings(bucket, "d"); len(w) != 0 {
			t.Errorf("cut at %d: %d warnings, want none: %v", n, len(w), w)
		}
	}
}

// TestN23DictForgedEntry (R-DEC-TS-NAGO-DICT, L-N23-36, Antwort „Texte maskieren (Empfohlen)“): a new value F, whose
// text holds a whole intact entry of N23 with an id of its own (the next id, the one of F), a salt of 0 and its CRC
// (over kind .. salt and the text as it would be stored unmasked, enumdict.go), behind the entries of A to D; then the
// CRC of the entry of F
// damaged, so that nago looks for the next intact entry behind it, through the text of F. After a reopen no point
// reads as the text of the forged entry, and enum.dict holds neither the text of F nor the forged text as written.
// Counter row: the same entry of F appended to the image in the format before N23 stands in it as written, before
// the rewrite.
func TestN23DictForgedEntry(t *testing.T) {
	h := n22cCapture(t)
	lay := n23DictLayout(t)
	const bucket = "dictforged"
	dir := t.TempDir()
	db := openDB(t, dir, n23DictOptions())
	c := column(t, db, bucket, "d", tsdb.Schema{Scheme: tsdb.SchemeEnum})
	sealed, record, wal := n23DictPoints()
	for _, p := range slices.Concat(sealed, record, []n22Point{wal}) {
		must(t, c.PutString(p.ts, p.str))
	}
	must(t, c.Sync())
	path := filepath.Join(dir, bucket, "d", "enum.dict")
	b, err := os.ReadFile(path)
	must(t, err)
	var idF uint32
	for id := range lay.entries(b) {
		idF = max(idF, id+1)
	}
	// the forged entry by the description: marker, kind of a text, id, length, a salt of 0, the CRC over kind .. salt and
	// the text as it would stand in the file, the text
	fake := "FAKE-" + strings.Repeat("z", 100)
	forged := slices.Concat(lay.marker, []byte{lay.textKind}, binary.LittleEndian.AppendUint32(nil, idF), binary.LittleEndian.AppendUint32(nil, uint32(len(fake))), make([]byte, 8))
	sum := crc32.ChecksumIEEE(slices.Concat(forged[4:], []byte(fake)))
	forged = slices.Concat(forged, binary.LittleEndian.AppendUint32(nil, sum), []byte(fake))
	if len(forged) != n23DictEntryHdr+len(fake) {
		t.Fatalf("premise: the forged entry has %d B, want %d", len(forged), n23DictEntryHdr+len(fake))
	}
	f := "F-" + string(forged) + "-F"
	start := len(b)
	must(t, c.PutString(30000, f))
	must(t, c.Sync())
	closeDB(t, db)
	b, err = os.ReadFile(path)
	must(t, err)
	if len(b) < start+n23DictEntryHdr || !bytes.Equal(b[start:start+4], lay.marker) || binary.LittleEndian.Uint32(b[start+5:]) != idF ||
		int(binary.LittleEndian.Uint32(b[start+9:])) != len(f) {
		t.Fatalf("premise: the entry of F at %d is not a text entry with the id %d and the length %d", start, idF, len(f))
	}
	b[start+n23DictCRCOff] ^= 0x5A // the CRC of the entry of F
	must(t, os.WriteFile(path, b, 0o644))

	db = openDB(t, dir, n23DictOptions())
	got := n23ReadEnum(t, lookup(t, db, bucket, "d"))
	closeDB(t, db)
	for ts, v := range got {
		if v == fake {
			t.Errorf("the point %d reads as the text of the forged entry", ts)
		}
	}
	want := n23DictModel()
	for ts, name := range want {
		if got[ts] != n23Value(name) {
			t.Errorf("the point %d of %s reads %q", ts, name, n23Short(got[ts]))
		}
	}
	if v, ok := got[30000]; ok && v != f {
		t.Errorf("the point of F reads %q, want F or missing", n23Short(v))
	}
	now, err := os.ReadFile(path)
	must(t, err)
	if bytes.Contains(now, []byte(f)) || bytes.Contains(now, []byte(fake)) {
		t.Errorf("enum.dict holds the text of F or the forged text as written")
	}
	if w := h.warnings(bucket, "d"); len(w) == 0 {
		t.Errorf("the damaged entry of F is not reported")
	}

	t.Run("counter row of the instrument: the forged entry, planted in place of the entry of F, is an entry nago reads", func(t *testing.T) {
		const bucket = "dictforgedplant"
		dir := t.TempDir()
		db := openDB(t, dir, n23DictOptions())
		c := column(t, db, bucket, "d", tsdb.Schema{Scheme: tsdb.SchemeEnum})
		for _, p := range slices.Concat(sealed, record, []n22Point{wal}) {
			must(t, c.PutString(p.ts, p.str))
		}
		must(t, c.Sync())
		path := filepath.Join(dir, bucket, "d", "enum.dict")
		before, err := os.ReadFile(path)
		must(t, err)
		must(t, c.PutString(30000, f))
		must(t, c.Sync())
		closeDB(t, db)
		if b, err := os.ReadFile(path); err != nil || binary.LittleEndian.Uint32(b[len(before)+5:]) != idF {
			t.Fatalf("premise: F has not the id %d here", idF)
		}
		must(t, os.WriteFile(path, slices.Concat(before, forged), 0o644))
		db = openDB(t, dir, n23DictOptions())
		got := n23ReadEnum(t, lookup(t, db, bucket, "d"))
		closeDB(t, db)
		if _, ok := got[30000]; !ok {
			t.Errorf("the planted forged entry is no entry nago reads: the point of its id is missing, so the row above would not see one taken from a text")
		}
	})
	t.Run("counter row: the entry of F in the image in the format before N23 stands as written, before the rewrite", func(t *testing.T) {
		img, err := os.ReadFile(filepath.Join(n23FixtureRoot, "db", "n23", "dict", "enum.dict"))
		must(t, err)
		old := slices.Concat(img, binary.LittleEndian.AppendUint32(nil, uint32(len(f))), binary.LittleEndian.AppendUint32(nil, crc32.ChecksumIEEE([]byte(f))), []byte(f))
		if !bytes.Contains(old, []byte(f)) || !bytes.Contains(old, []byte(fake)) {
			t.Errorf("the image does not hold the text of F as written")
		}
		if _, texts := n23OldDictEntries(old); !slices.Contains(texts, f) {
			t.Errorf("the format before N23 does not read F from the image")
		}
	})
}
