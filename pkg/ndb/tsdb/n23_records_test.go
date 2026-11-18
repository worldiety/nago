// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The acceptance tests of R-DEC-TS-NAGO-RECORDS (wokoda N23, L-N23-6) and of the general rows of
// R-DEC-TS-NAGO-DAMAGE: a damaged continuation record is reported before its values go and costs its own values only.
// The tests write a column with the three records S1, S2 and S3 in the open block (points, Sync, points, Sync, points,
// Sync) with random values, so that a byte in the middle of a record meets values, and damage the open chunk with os,
// after a crash (a crash copy, without Close) or while the DB is open. Where a record begins and ends they observe from
// the size of the open chunk before and after each Sync; the layout of a record (| Marker 8 | BodyLen 4 | Kind 1 |
// Count 4 | Body | CRC32 4 |) is the one of chunk.go:295–305, which N23 does not change.

const (
	n23RecPoints = 10 // points of a record
	n23RecBlock  = 64 // points of a block: three records do not fill it
)

func n23RecOptions() tsdb.Options { return quietOptions(n23RecBlock) }

// n23RecSpan is a column with three records in the open block.
type n23RecSpan struct {
	dir, bucket string
	db          *tsdb.DB // the live DB, nil after a crash copy
	col         *tsdb.Column
	chunk       string   // the path of the open chunk
	recs        [][2]int // S1, S2, S3: start and end
	model       map[int64]string
	r           *rand.Rand
	next        int64 // the time of the next point
}

// n23RecWrite writes the column bucket/c with S1, S2 and S3 into a new database; crash leaves a crash copy and no
// live DB.
func n23RecWrite(t *testing.T, bucket string, crash bool) *n23RecSpan {
	t.Helper()
	dir := t.TempDir()
	s := &n23RecSpan{dir: dir, bucket: bucket, model: map[int64]string{}, r: rand.New(rand.NewPCG(23, 6)), next: 1000}
	s.db = openDB(t, dir, n23RecOptions())
	s.col = column(t, s.db, bucket, "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
	for range 3 {
		s.puts(t, n23RecPoints)
		if s.chunk == "" {
			s.chunk = n23OpenChunk(t, filepath.Join(dir, bucket, "c"))
		}
		start := n23Size(t, s.chunk)
		must(t, s.col.Sync())
		s.recs = append(s.recs, [2]int{int(start), int(n23Size(t, s.chunk))})
	}
	for i := 1; i < 3; i++ {
		if s.recs[i][0] != s.recs[i-1][1] {
			t.Fatalf("premise: the record S%d begins at %d, S%d ended at %d", i+1, s.recs[i][0], i, s.recs[i-1][1])
		}
	}
	if crash {
		s.dir = crashCopy(t, dir)
		s.chunk = filepath.Join(s.dir, bucket, "c", filepath.Base(s.chunk))
		closeDB(t, s.db)
		s.db, s.col = nil, nil
	}
	return s
}

// puts appends n points with random values.
func (s *n23RecSpan) puts(t *testing.T, n int) {
	t.Helper()
	for range n {
		v := s.r.Int64N(1_000_000_000)
		must(t, s.col.PutI64(s.next, v))
		s.model[s.next] = strconv.FormatInt(v, 10)
		s.next += 1000
	}
}

// without is the model without the points of the records of index in.
func (s *n23RecSpan) without(in ...int) map[int64]string {
	m := cloneMap(s.model)
	for _, i := range in {
		for k := range n23RecPoints {
			delete(m, 1000+int64(i*n23RecPoints+k)*1000)
		}
	}
	return m
}

// n23OpenChunk is the open chunk of the column directory.
func n23OpenChunk(t *testing.T, dir string) string {
	t.Helper()
	m, _ := filepath.Glob(filepath.Join(dir, "*_.tsb"))
	if len(m) != 1 {
		t.Fatalf("premise: %d open chunks in %s", len(m), dir)
	}
	return m[0]
}

// edit applies fn to the open chunk.
func (s *n23RecSpan) edit(t *testing.T, fn func(b []byte) []byte) {
	t.Helper()
	b, err := os.ReadFile(s.chunk)
	must(t, err)
	must(t, os.WriteFile(s.chunk, fn(b), 0o644))
}

// n23RecBody is a byte in the middle of the body of record i.
func (s *n23RecSpan) body(i int) int { return (s.recs[i][0] + 17 + s.recs[i][1] - 4) / 2 }

// n23RecSetLen sets the BodyLen of record i to n.
func (s *n23RecSpan) setLen(b []byte, i, n int) []byte {
	binary.LittleEndian.PutUint32(b[s.recs[i][0]+8:], uint32(n))
	return b
}

func (s *n23RecSpan) bodyLen(b []byte, i int) int {
	return int(binary.LittleEndian.Uint32(b[s.recs[i][0]+8:]))
}

// n23RecReports checks the warnings of the column since from: their number and their offsets in the open chunk.
func n23RecReports(t *testing.T, h *n22cHandler, when, bucket, chunk string, from int, offs []int64) {
	t.Helper()
	recs := n23Warnings(h, bucket, "c", from)
	if len(recs) != len(offs) {
		t.Errorf("%s: %d warnings of the column, want %d: %v", when, len(recs), len(offs), recs)
		return
	}
	got := n23Offsets(t, when, recs, filepath.Base(chunk))
	slices.Sort(got)
	if want := slices.Sorted(slices.Values(offs)); !slices.Equal(got, want) {
		t.Errorf("%s: the reports have the offsets %v, want %v", when, got, want)
	}
}

// TestN23RecordsDamage is the table of R-DEC-TS-NAGO-RECORDS over an open after a crash, with the rows "zwei
// beschädigte Rahmen" and "ein genullter Bereich über mehrere Rahmen" of R-DEC-TS-NAGO-DAMAGE.
func TestN23RecordsDamage(t *testing.T) {
	h := n22cCapture(t)
	type row struct {
		name    string
		edit    func(s *n23RecSpan, b []byte) []byte
		lost    []int // the records whose points are missing
		reports []int // the records whose start a report names
		again   bool  // a second crash and open, which reports nothing more
	}
	rows := []row{
		{name: "counter row: intact records"},
		{name: "a byte in the body of S2", edit: func(s *n23RecSpan, b []byte) []byte { return n23Flip(b, s.body(1)) },
			lost: []int{1}, reports: []int{1}},
		{name: "a byte in the body of S2, then a crash again and a second open", edit: func(s *n23RecSpan, b []byte) []byte { return n23Flip(b, s.body(1)) },
			lost: []int{1}, reports: []int{1}, again: true},
		{name: "a byte in the body of S1 (the first record)", edit: func(s *n23RecSpan, b []byte) []byte { return n23Flip(b, s.body(0)) },
			lost: []int{0}, reports: []int{0}},
		{name: "a byte in the body of S3 (the last record, with its full length)", edit: func(s *n23RecSpan, b []byte) []byte { return n23Flip(b, s.body(2)) },
			lost: []int{2}, reports: []int{2}},
		{name: "a byte in the marker of S1", edit: func(s *n23RecSpan, b []byte) []byte { return n23Flip(b, s.recs[0][0]+3) },
			lost: []int{0}, reports: []int{0}},
		{name: "a byte in the marker of S2", edit: func(s *n23RecSpan, b []byte) []byte { return n23Flip(b, s.recs[1][0]+3) },
			lost: []int{1}, reports: []int{1}},
		{name: "a byte in the marker of S3", edit: func(s *n23RecSpan, b []byte) []byte { return n23Flip(b, s.recs[2][0]+3) },
			lost: []int{2}, reports: []int{2}},
		{name: "S2 with its marker zeroed", edit: func(s *n23RecSpan, b []byte) []byte { clear(b[s.recs[1][0]:s.recs[1][1]]); return b },
			lost: []int{1}, reports: []int{1}},
		{name: "the length of S2 damaged, so that S2 ends before S3", edit: func(s *n23RecSpan, b []byte) []byte { return s.setLen(b, 1, s.bodyLen(b, 1)-3) },
			lost: []int{1}, reports: []int{1}},
		{name: "the length of S2 damaged, so that S2 ends behind the start of S3, before the end of the file",
			edit: func(s *n23RecSpan, b []byte) []byte {
				return s.setLen(b, 1, s.bodyLen(b, 1)+(s.recs[2][1]-s.recs[2][0])/2)
			},
			lost: []int{1}, reports: []int{1}},
		{name: "the length of S3 damaged, so that S3 reaches over the end of the file (limit: like a torn end)",
			edit: func(s *n23RecSpan, b []byte) []byte { return s.setLen(b, 2, s.bodyLen(b, 2)+100) },
			lost: []int{2}},
		{name: "the file cut in the middle of S3 (a torn end)", edit: func(s *n23RecSpan, b []byte) []byte { return b[:(s.recs[2][0]+s.recs[2][1])/2] },
			lost: []int{2}},
		{name: "the file cut in the middle of S2, S3 gone (limit: everything behind the cut is missing)",
			edit: func(s *n23RecSpan, b []byte) []byte { return b[:(s.recs[1][0]+s.recs[1][1])/2] },
			lost: []int{1, 2}},
		{name: "two damaged records at different places", edit: func(s *n23RecSpan, b []byte) []byte {
			return n23Flip(n23Flip(b, s.body(0)), s.body(2))
		}, lost: []int{0, 2}, reports: []int{0, 2}},
		{name: "a zeroed region over two records", edit: func(s *n23RecSpan, b []byte) []byte { clear(b[s.recs[0][0]:s.recs[1][1]]); return b },
			lost: []int{0, 1}, reports: []int{0}},
	}
	for i, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			bucket := "records" + strconv.Itoa(i)
			s := n23RecWrite(t, bucket, true)
			if r.edit != nil {
				s.edit(t, func(b []byte) []byte { return r.edit(s, b) })
			}
			var offs []int64
			for _, k := range r.reports {
				offs = append(offs, int64(s.recs[k][0]))
			}
			want := s.without(r.lost...)
			db := openDB(t, s.dir, n23RecOptions())
			c := lookup(t, db, bucket, "c")
			n23CheckDec(t, "the first read", n23ReadDec(t, c), want)
			n23CheckDec(t, "the second read", n23ReadDec(t, c), want)
			n23RecReports(t, h, "the open after the crash", bucket, s.chunk, 0, offs)
			if !r.again {
				closeDB(t, db)
				return
			}
			crashed := crashCopy(t, s.dir)
			_ = db.Close()
			mark := len(h.warnings(bucket, "c"))
			db = openDB(t, crashed, n23RecOptions())
			n23CheckDec(t, "after the second crash", n23ReadDec(t, lookup(t, db, bucket, "c")), want)
			closeDB(t, db)
			n23RecReports(t, h, "the second open", bucket, s.chunk, mark, nil)
		})
	}
}

// TestN23RecordsDamageLive: the rows of R-DEC-TS-NAGO-RECORDS with the DB open: a byte in the body of S2, then a read,
// the block filled (sealed) or a compaction, which rewrites the open chunk; each report comes before the record goes,
// and after a reopen none. The variants without a read before show, that the seal and the compaction report it
// themselves.
func TestN23RecordsDamageLive(t *testing.T) {
	h := n22cCapture(t)
	for i, r := range []struct {
		name string
		read bool   // a read after the damage
		then string // "", "seal" or "compact"
	}{
		{name: "a byte in the body of S2, then a read", read: true},
		{name: "the same, then the block filled and sealed, a reopen", read: true, then: "seal"},
		{name: "the block filled and sealed without a read before, a reopen", then: "seal"},
		{name: "the same with a compaction, which rewrites the open chunk, instead of the seal", read: true, then: "compact"},
		{name: "a compaction without a read before, a reopen", then: "compact"},
	} {
		t.Run(r.name, func(t *testing.T) {
			bucket := "recordslive" + strconv.Itoa(i)
			s := n23RecWrite(t, bucket, false)
			s.edit(t, func(b []byte) []byte { return n23Flip(b, s.body(1)) })
			off := []int64{int64(s.recs[1][0])}
			if r.read {
				n23CheckDec(t, "the read", n23ReadDec(t, s.col), s.without(1))
				n23RecReports(t, h, "after the read", bucket, s.chunk, 0, off)
			}
			switch r.then {
			case "seal":
				before := n23Size(t, s.chunk)
				s.puts(t, n23RecBlock-3*n23RecPoints)
				if n23Size(t, s.chunk) == before {
					t.Fatalf("premise: the block did not seal, the open chunk did not change")
				}
			case "compact":
				// a correction of a point of S1 into the head, then Flush: the compaction rewrites the open chunk
				must(t, s.col.PutI64(1000, 42))
				s.model[1000] = "42"
				must(t, s.col.Flush())
			}
			want := s.without(1)
			if r.then != "" {
				n23RecReports(t, h, "before the "+r.then, bucket, s.chunk, 0, off)
			}
			closeDB(t, s.db)
			if r.then == "" {
				return
			}
			mark := len(h.warnings(bucket, "c"))
			db := openDB(t, s.dir, n23RecOptions())
			n23CheckDec(t, "after the reopen", n23ReadDec(t, lookup(t, db, bucket, "c")), want)
			closeDB(t, db)
			if w := n23Warnings(h, bucket, "c", mark); len(w) != 0 {
				t.Errorf("after the %s and a reopen: %d warnings, want none: %v", r.then, len(w), w)
			}
		})
	}
}

// TestN23CutAnywhereNoReport (R-DEC-TS-NAGO-DAMAGE, rows two and three of "Absturz und Beschädigung"): what only
// appends (continuation records, records of the WAL, entries of the dictionary) is cut at every length between its
// size before and after a step, as a crash in the middle of the step leaves it; after the open no report comes, the
// torn frame is missing and the frames before it hold. Then a new frame, shorter than the torn one, a crash again, an
// open: no report, every frame but the torn one holds, and no rest of the torn one is read as a frame; in the
// dictionary the new value does not get the id of the torn entry. Counter row: the same frame with its full length
// and a damaged byte in its body is reported once. With -short a few lengths only.
func TestN23CutAnywhereNoReport(t *testing.T) {
	h := n22cCapture(t)
	lengths := func(from, to int) []int {
		var res []int
		for n := from; n < to; n++ {
			if !testing.Short() || n == from || n == from+1 || n == (from+to)/2 || n == to-1 {
				res = append(res, n)
			}
		}
		return res
	}
	type step struct {
		name string
		// write writes the column bucket/c in the database at dir up to the step under test and returns the file it
		// appends to, the model before it and the time and value of the frame it appends; the frame itself is written by
		// frame, which returns the model after it
		setup  func(t *testing.T, dir, bucket string) (c *tsdb.Column, file string, before map[int64]string, s surfaceScheme)
		frame  func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string
		short  func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string // a shorter frame after the cut
		damage func(b []byte, from, to int) []byte                                     // the counter row
	}
	big := int64(1) << 50
	steps := []step{
		{
			name: "a continuation record",
			setup: func(t *testing.T, dir, bucket string) (*tsdb.Column, string, map[int64]string, surfaceScheme) {
				db := openDB(t, dir, n23RecOptions())
				c := column(t, db, bucket, "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				m := map[int64]string{}
				for ts := int64(1000); ts <= 5000; ts += 1000 {
					must(t, c.PutI64(ts, ts*7))
					m[ts] = strconv.FormatInt(ts*7, 10)
				}
				must(t, c.Sync())
				return c, n23OpenChunk(t, filepath.Join(dir, bucket, "c")), m, decimalScheme()
			},
			frame: func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string {
				m = cloneMap(m)
				for ts := int64(6000); ts <= 12000; ts += 1000 {
					must(t, c.PutI64(ts, big+ts))
					m[ts] = strconv.FormatInt(big+ts, 10)
				}
				must(t, c.Sync())
				return m
			},
			short: func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string {
				m = cloneMap(m)
				must(t, c.PutI64(20000, 1))
				m[20000] = "1"
				must(t, c.Sync())
				return m
			},
			damage: func(b []byte, from, to int) []byte { return n23Flip(b, (from+17+to-4)/2) },
		},
		{
			name: "a record of the WAL",
			setup: func(t *testing.T, dir, bucket string) (*tsdb.Column, string, map[int64]string, surfaceScheme) {
				db := openDB(t, dir, n23WALOptions())
				c := column(t, db, bucket, "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				m := map[int64]string{}
				for _, p := range n23WALBase() {
					must(t, c.PutI64(p.ts, p.v))
					m[p.ts] = strconv.FormatInt(p.v, 10)
				}
				must(t, c.Sync())
				must(t, c.PutI64(n23K1, 111))
				m[n23K1] = "111"
				return c, filepath.Join(dir, bucket, "c", "head.wal"), m, decimalScheme()
			},
			frame: func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string {
				m = cloneMap(m)
				must(t, c.PutI64(n23K2, big))
				m[n23K2] = strconv.FormatInt(big, 10)
				return m
			},
			short: func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string {
				m = cloneMap(m)
				must(t, c.Delete(n23K3))
				delete(m, n23K3)
				return m
			},
			damage: func(b []byte, from, to int) []byte { return n23Flip(b, to-1) },
		},
		{
			name: "the first record of an empty WAL, with its marker",
			setup: func(t *testing.T, dir, bucket string) (*tsdb.Column, string, map[int64]string, surfaceScheme) {
				db := openDB(t, dir, n23WALOptions())
				c := column(t, db, bucket, "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				m := map[int64]string{}
				for _, p := range n23WALBase() {
					must(t, c.PutI64(p.ts, p.v))
					m[p.ts] = strconv.FormatInt(p.v, 10)
				}
				must(t, c.Sync())
				return c, filepath.Join(dir, bucket, "c", "head.wal"), m, decimalScheme()
			},
			frame: func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string {
				m = cloneMap(m)
				must(t, c.PutI64(n23K2, big))
				m[n23K2] = strconv.FormatInt(big, 10)
				return m
			},
			short: func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string {
				m = cloneMap(m)
				must(t, c.Delete(n23K3))
				delete(m, n23K3)
				return m
			},
			damage: func(b []byte, from, to int) []byte { return n23Flip(b, to-1) },
		},
		{
			name: "an entry of the dictionary",
			setup: func(t *testing.T, dir, bucket string) (*tsdb.Column, string, map[int64]string, surfaceScheme) {
				db := openDB(t, dir, n23DictOptions())
				c := column(t, db, bucket, "c", tsdb.Schema{Scheme: tsdb.SchemeEnum})
				m := map[int64]string{}
				for i, name := range []string{"A", "B"} {
					ts := int64(i+1) * 1000
					must(t, c.PutString(ts, n23Value(name)))
					m[ts] = n23Value(name)
				}
				must(t, c.Sync())
				return c, filepath.Join(dir, bucket, "c", "enum.dict"), m, enumScheme()
			},
			// a new value, whose point stays in memory: a crash in the middle of the step leaves no point of it
			frame: func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string {
				must(t, c.PutString(3000, n23Value("D")))
				return cloneMap(m) // its point is only in memory, a crash loses it
			},
			short: func(t *testing.T, c *tsdb.Column, m map[int64]string) map[int64]string {
				m = cloneMap(m)
				must(t, c.PutString(4000, "kurz"))
				m[4000] = "kurz"
				must(t, c.Sync())
				return m
			},
			damage: func(b []byte, from, to int) []byte { return n23Flip(b, (from+to)/2) },
		},
	}
	for si, st := range steps {
		t.Run(st.name, func(t *testing.T) {
			bucket := "cut" + strconv.Itoa(si)
			live := t.TempDir()
			c, file, before, scheme := st.setup(t, live, bucket)
			colDir := filepath.Join(live, bucket, "c")
			snapBefore := readSnapshot(t, colDir)
			after := st.frame(t, c, before)
			snapAfter := readSnapshot(t, colDir)
			name := filepath.Base(file)
			from, to := len(snapBefore[name]), len(snapAfter[name])
			if to <= from {
				t.Fatalf("premise: the step did not append to %s (%d to %d bytes)", name, from, to)
			}
			for other, b := range snapAfter {
				if other != name && string(b) != string(snapBefore[other]) && !strings.HasSuffix(other, "_.tsb") {
					t.Logf("the step also changed %s", other)
				}
			}
			open := func(t *testing.T, state snapshot) (string, *tsdb.DB, *tsdb.Column) {
				dir := filepath.Join(t.TempDir(), "db")
				cd := filepath.Join(dir, bucket, "c")
				must(t, os.MkdirAll(cd, 0o755))
				for n, b := range state {
					must(t, os.WriteFile(filepath.Join(cd, n), b, 0o644))
				}
				opts := n23RecOptions()
				db, err := tsdb.Open(dir, opts)
				must(t, err)
				return dir, db, lookup(t, db, bucket, "c")
			}
			// the step's other files as after it, which only matter for a record of the WAL with a new entry: none here
			for _, n := range lengths(from, to) {
				state := cloneSnapshot(snapBefore)
				for other, b := range snapAfter {
					if other != name {
						state[other] = b
					}
				}
				state[name] = snapAfter[name][:n]
				mark := len(h.warnings(bucket, "c"))
				dir, db, col := open(t, state)
				got, err := scanMap(scheme, col)
				must(t, err)
				n23CheckMapNames(t, fmt.Sprintf("cut at %d of %d", n, to), got, before)
				// a shorter frame, a crash again, an open
				m := st.short(t, col, before)
				again := crashCopy(t, dir)
				_ = db.Close()
				db = openDB(t, again, n23RecOptions())
				got, err = scanMap(scheme, lookup(t, db, bucket, "c"))
				must(t, err)
				n23CheckMapNames(t, fmt.Sprintf("cut at %d, then a shorter frame and a crash", n), got, m)
				closeDB(t, db)
				if w := n23Warnings(h, bucket, "c", mark); len(w) != 0 {
					t.Errorf("cut at %d of %d: %d warnings, want none: %v", n, to, len(w), w)
				}
			}
			// the state after the step reads the frame, without a report
			mark := len(h.warnings(bucket, "c"))
			_, db, col := open(t, snapAfter)
			got, err := scanMap(scheme, col)
			must(t, err)
			n23CheckMapNames(t, "the state after the step", got, after)
			_ = db.Close()
			if w := n23Warnings(h, bucket, "c", mark); len(w) != 0 {
				t.Errorf("the state after the step: %d warnings, want none: %v", len(w), w)
			}
			// counter row: the frame with its full length and a damaged byte in its body
			state := cloneSnapshot(snapAfter)
			state[name] = st.damage(slices.Clone(snapAfter[name]), from, to)
			mark = len(h.warnings(bucket, "c"))
			_, db, col = open(t, state)
			if _, err := scanMap(scheme, col); err != nil {
				t.Errorf("counter row: the read fails: %v", err)
			}
			_ = db.Close()
			if w := n23Warnings(h, bucket, "c", mark); len(w) != 1 {
				t.Errorf("counter row: %d warnings, want one: %v", len(w), w)
			}
		})
	}
}

// n23CheckMapNames compares two maps of points and names enum texts by n23NameOf in a message.
func n23CheckMapNames(t *testing.T, when string, got, want map[int64]string) {
	t.Helper()
	if !mapsEqual(got, want) {
		t.Errorf("%s: reads %v, want %v", when, n23Points(got), n23Points(want))
	}
}
