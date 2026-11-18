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
	"math/bits"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The acceptance tests of R-DEC-TS-NAGO-WAL (wokoda N23, L-N23-7, L-N23-25, L-N23-32, L-N23-33): a record of the WAL is
// a correction or a deletion out of order. The tests write a column with sealed points (n23WALBase), then out of order
// the corrections K1, K2 and K3 at three times and the deletion L of a fourth, in this order, end the process without
// Close (a crash copy: a Close folds a read head into the chunks), damage head.wal with os and open again.
//
// Where a record begins and ends the tests observe from the size of head.wal after each step; the layout of a record
// of N23 (kind, length, payload) they take from the description of the format in the code (n23WALLayout), the one of
// 8ae1daaee from head.go:73–86 of that stand. "Gilt" means: a read yields the value of the correction at its time
// exactly once, or no point at the time of the deletion.

// n23WALSpan is the column of a WAL test: its database after the crash, and the bounds of the marker and the records.
type n23WALSpan struct {
	dir, bucket string
	path        string   // head.wal
	marker      int      // the end of the marker (0 for a WAL in the format before N23)
	recs        [][2]int // K1, K2, K3, L: start and end
}

// n23WALNew writes column bucket/w with the code under test: n23WALBase, Close, K1 … L (n23WALSteps), a crash copy.
func n23WALNew(t *testing.T, bucket string) n23WALSpan {
	t.Helper()
	dir, ends := n23WriteWAL(t, t.TempDir(), bucket, "w", 4)
	s := n23WALSpan{dir: dir, bucket: bucket, path: filepath.Join(dir, bucket, "w", "head.wal")}
	b, err := os.ReadFile(s.path)
	must(t, err)
	kind, mEnd, intact := n23OldFirstFrame(b)
	if !intact || kind == 0 || kind == 1 {
		t.Fatalf("premise: head.wal does not begin with a marker, which the layout of 8ae1daaee reads as an intact frame of a kind but 0 and 1 (kind %d, intact %v)", kind, intact)
	}
	if int64(len(b)) != ends[3] {
		t.Fatalf("premise: head.wal has %d bytes, %d after L", len(b), ends[3])
	}
	if k1 := ends[1] - ends[0]; int64(mEnd)+k1 != ends[0] {
		t.Fatalf("premise: the marker ends at %d, K1 (as long as K2, %d bytes) at %d", mEnd, k1, ends[0])
	}
	s.marker = mEnd
	starts := []int{mEnd, int(ends[0]), int(ends[1]), int(ends[2])}
	for i := range 4 {
		s.recs = append(s.recs, [2]int{starts[i], int(ends[i])})
	}
	return s
}

// n23WALOld copies the image col of testdata/8ae1daaee to the column bucket/w of a new database.
func n23WALOld(t *testing.T, col, bucket string) n23WALSpan {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "db")
	must(t, os.CopyFS(filepath.Join(dir, bucket, "w"), os.DirFS(filepath.Join(n23FixtureRoot, "db", "n23", col))))
	s := n23WALSpan{dir: dir, bucket: bucket, path: filepath.Join(dir, bucket, "w", "head.wal")}
	b, err := os.ReadFile(s.path)
	must(t, err)
	starts, end := n22WALRecords(b)
	if len(starts) != 4 || end != len(b) {
		t.Fatalf("premise: the image holds %d intact records up to %d of %d bytes, want K1 to L", len(starts), end, len(b))
	}
	for i, at := range starts {
		e := len(b)
		if i+1 < len(starts) {
			e = starts[i+1]
		}
		s.recs = append(s.recs, [2]int{at, e})
	}
	return s
}

// edit applies fn to head.wal.
func (s n23WALSpan) edit(t *testing.T, fn func(b []byte) []byte) {
	t.Helper()
	b, err := os.ReadFile(s.path)
	must(t, err)
	must(t, os.WriteFile(s.path, fn(b), 0o644))
}

// n23WALWant is what column w reads: n23WALBase with the steps of n23WALSteps, whose index is in apply (0 K1, 1 K2,
// 2 K3, 3 L), and the points of extra.
func n23WALWant(apply []int, extra map[int64]int64) map[int64]string {
	m := n23WALModel(0)
	for _, i := range apply {
		st := n23WALSteps[i]
		if st.del {
			delete(m, st.ts)
		} else {
			m[st.ts] = strconv.FormatInt(st.v, 10)
		}
	}
	for ts, v := range extra {
		m[ts] = strconv.FormatInt(v, 10)
	}
	return m
}

// n23ReadDec reads every point of the decimal column c.
func n23ReadDec(t *testing.T, c *tsdb.Column) map[int64]string {
	t.Helper()
	got, err := scanMap(decimalScheme(), c)
	if err != nil {
		t.Fatalf("ScanI64: %v", err)
	}
	return got
}

func n23CheckDec(t *testing.T, when string, got, want map[int64]string) {
	t.Helper()
	if !mapsEqual(got, want) {
		t.Errorf("%s: reads %v, want %v", when, describeMap(got), describeMap(want))
	}
}

// n23Reports checks the warnings of the bucket and column w since from: their number and, if offs is not nil, their
// offsets in head.wal.
func n23Reports(t *testing.T, h *n22cHandler, when, bucket string, from int, offs []int64) {
	t.Helper()
	recs := n23Warnings(h, bucket, "w", from)
	if len(recs) != len(offs) {
		t.Errorf("%s: %d warnings of the column, want %d: %v", when, len(recs), len(offs), recs)
		return
	}
	got := n23Offsets(t, when, recs, "head.wal")
	slices.Sort(got)
	want := slices.Sorted(slices.Values(offs))
	for i := range want {
		if want[i] >= 0 && got[i] != want[i] {
			t.Errorf("%s: the reports have the offsets %v, want %v", when, got, want)
			return
		}
	}
}

// n23Open opens the database of the span and returns column w.
func n23Open(t *testing.T, dir, bucket string) (*tsdb.DB, *tsdb.Column) {
	t.Helper()
	db := openDB(t, dir, n23WALOptions())
	return db, lookup(t, db, bucket, "w")
}

// TestN23WALDamage is the table "WAL im Format von N23" of R-DEC-TS-NAGO-WAL over a WAL, which the code under test
// wrote. Each row is one of damage beside the counter row, an intact WAL.
func TestN23WALDamage(t *testing.T) {
	h := n22cCapture(t)
	type row struct {
		name  string
		edit  func(t *testing.T, s n23WALSpan, b []byte) []byte
		apply []int   // the steps that hold
		offs  []int64 // the offsets of the reports: -1 any; nil none
		at    func(s n23WALSpan) []int64
	}
	payloadByte := func(t *testing.T, s n23WALSpan, b []byte, i int) []byte {
		l := n23WALLayout(t)
		at := s.recs[i][0] + l.payloadOff
		if at >= s.recs[i][1] {
			t.Fatalf("premise: the record %d has no payload", i)
		}
		return n23Flip(b, at)
	}
	setLen := func(t *testing.T, s n23WALSpan, b []byte, i, n int) []byte {
		l := n23WALLayout(t)
		at := s.recs[i][0] + l.lenOff
		switch l.lenSize {
		case 4:
			binary.LittleEndian.PutUint32(b[at:], uint32(n))
		case 2:
			binary.LittleEndian.PutUint16(b[at:], uint16(n))
		default:
			t.Fatalf("premise: a length of %d bytes", l.lenSize)
		}
		return b
	}
	lenOf := func(t *testing.T, s n23WALSpan, b []byte, i int) int {
		l := n23WALLayout(t)
		at := s.recs[i][0] + l.lenOff
		if l.lenSize == 2 {
			return int(binary.LittleEndian.Uint16(b[at:]))
		}
		return int(binary.LittleEndian.Uint32(b[at:]))
	}
	recAt := func(i int) func(s n23WALSpan) []int64 {
		return func(s n23WALSpan) []int64 { return []int64{int64(s.recs[i][0])} }
	}
	rows := []row{
		{name: "counter row: an intact WAL", apply: []int{0, 1, 2, 3}},
		{name: "a byte in the payload of K2", apply: []int{0, 2, 3}, at: recAt(1),
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte { return payloadByte(t, s, b, 1) }},
		{name: "a bit of the kind of K2 flipped, so that the correction would become a deletion", apply: []int{0, 2, 3}, at: recAt(1),
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte {
				l := n23WALLayout(t)
				k2, del := b[s.recs[1][0]+l.kindOff], b[s.recs[3][0]+l.kindOff]
				if bits.OnesCount8(k2^del) != 1 {
					t.Fatalf("premise: the kinds of a correction (%#x) and of a deletion (%#x) differ in more than a bit", k2, del)
				}
				b[s.recs[1][0]+l.kindOff] = del
				return b
			}},
		{name: "the kind of K2 set to a value nago does not know", apply: []int{0, 2, 3}, at: recAt(1),
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte {
				l := n23WALLayout(t)
				if k := b[s.recs[1][0]+l.kindOff]; k == n22UnknownKind || b[s.recs[3][0]+l.kindOff] == n22UnknownKind {
					t.Fatalf("premise: the kind %#x is known", n22UnknownKind)
				}
				b[s.recs[1][0]+l.kindOff] = n22UnknownKind
				return b
			}},
		{name: "a byte in the payload of L", apply: []int{0, 1, 2}, at: recAt(3),
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte { return payloadByte(t, s, b, 3) }},
		{name: "the length of K2 damaged, so that K2 ends before the end of the file", apply: []int{0, 2, 3}, at: recAt(1),
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte { return setLen(t, s, b, 1, lenOf(t, s, b, 1)-1) }},
		{name: "the length of K2 damaged, so that K2 reaches over the end of the file, K3 and L intact", apply: []int{0, 2, 3}, at: recAt(1),
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte { return setLen(t, s, b, 1, lenOf(t, s, b, 1)+len(b)) }},
		{name: "K2 with its header zeroed", apply: []int{0, 2, 3}, at: recAt(1),
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte { clear(b[s.recs[1][0]:s.recs[1][1]]); return b }},
		{name: "a byte in the marker", apply: []int{0, 1, 2, 3}, at: func(n23WALSpan) []int64 { return []int64{0} },
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte { return n23Flip(b, s.marker/2) }},
		{name: "the file cut in the middle of L", apply: []int{0, 1, 2},
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte { return b[:(s.recs[3][0]+s.recs[3][1])/2] }},
		{name: "the file cut in the middle of K2, K3 and L gone", apply: []int{0},
			edit: func(t *testing.T, s n23WALSpan, b []byte) []byte { return b[:(s.recs[1][0]+s.recs[1][1])/2] }},
	}
	for i, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			bucket := "waldamage" + strconv.Itoa(i)
			s := n23WALNew(t, bucket)
			if r.edit != nil {
				s.edit(t, func(b []byte) []byte { return r.edit(t, s, b) })
			}
			var offs []int64
			if r.at != nil {
				offs = r.at(s)
			}
			for life := range 2 {
				mark := len(h.warnings(bucket, "w"))
				db, c := n23Open(t, s.dir, bucket)
				want := n23WALWant(r.apply, nil)
				n23CheckDec(t, fmt.Sprintf("life %d", life+1), n23ReadDec(t, c), want)
				n23CheckDec(t, fmt.Sprintf("life %d, read again", life+1), n23ReadDec(t, c), want)
				n23Reports(t, h, fmt.Sprintf("life %d", life+1), bucket, mark, offs)
				// a crash again, so that the head stays in head.wal: the second life opens the same WAL
				s.dir = crashCopy(t, s.dir)
				s.path = filepath.Join(s.dir, bucket, "w", "head.wal")
				_ = db.Close()
			}
		})
	}

	t.Run("cut in the middle of L, then a correction K4 shorter than L, a crash, open", func(t *testing.T) {
		const bucket = "walshort"
		// L deletes a point at a large time, so that a correction at a small time can be shorter
		dir := t.TempDir()
		db := openDB(t, dir, n23WALOptions())
		c := column(t, db, bucket, "w", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
		const far = int64(1) << 40
		model := n23WALModel(0)
		for _, p := range n23WALBase() {
			must(t, c.PutI64(p.ts, p.v))
		}
		must(t, c.PutI64(far, 99))
		model[far] = "99"
		closeDB(t, db)
		db = openDB(t, dir, n23WALOptions())
		c = lookup(t, db, bucket, "w")
		path := filepath.Join(dir, bucket, "w", "head.wal")
		var ends []int64
		for _, st := range n23WALSteps[:3] {
			must(t, c.PutI64(st.ts, st.v))
			model[st.ts] = strconv.FormatInt(st.v, 10)
			ends = append(ends, n23Size(t, path))
		}
		must(t, c.Delete(far))
		ends = append(ends, n23Size(t, path))
		crashed := crashCopy(t, dir)
		closeDB(t, db)
		lLen := ends[3] - ends[2]
		b, err := os.ReadFile(filepath.Join(crashed, bucket, "w", "head.wal"))
		must(t, err)
		must(t, os.WriteFile(filepath.Join(crashed, bucket, "w", "head.wal"), b[:ends[2]+lLen/2], 0o644))

		mark := len(h.warnings(bucket, "w"))
		db, c = n23Open(t, crashed, bucket)
		n23CheckDec(t, "after the cut", n23ReadDec(t, c), model)
		cutPath := filepath.Join(crashed, bucket, "w", "head.wal")
		before := n23Size(t, cutPath)
		must(t, c.PutI64(1000, 7)) // K4: a correction at a small time
		model[1000] = "7"
		if k4 := n23Size(t, cutPath) - before; k4 >= lLen {
			t.Fatalf("premise: the record of K4 has %d bytes, L %d: the format has no correction shorter than L", k4, lLen)
		}
		again := crashCopy(t, crashed)
		_ = db.Close()
		db, c = n23Open(t, again, bucket)
		n23CheckDec(t, "after K4 and a second crash", n23ReadDec(t, c), model)
		closeDB(t, db)
		n23Reports(t, h, "both lives", bucket, mark, nil)
	})

	t.Run("a byte in the payload of K2, then a correction K4, then a reopen", func(t *testing.T) {
		const bucket = "walk4"
		s := n23WALNew(t, bucket)
		s.edit(t, func(b []byte) []byte { return n23Flip(b, s.recs[1][0]+n23WALLayout(t).payloadOff) })
		want := n23WALWant([]int{0, 2, 3}, map[int64]int64{n23K4: 444})
		mark := len(h.warnings(bucket, "w"))
		db, c := n23Open(t, s.dir, bucket)
		must(t, c.PutI64(n23K4, 444))
		n23CheckDec(t, "after K4", n23ReadDec(t, c), want)
		n23Reports(t, h, "life 1", bucket, mark, []int64{int64(s.recs[1][0])})
		crashed := crashCopy(t, s.dir)
		_ = db.Close()
		mark = len(h.warnings(bucket, "w"))
		db, c = n23Open(t, crashed, bucket)
		n23CheckDec(t, "after reopening", n23ReadDec(t, c), want)
		n23Reports(t, h, "life 2", bucket, mark, []int64{int64(s.recs[1][0])})
		// the same, then a compaction (Close), then a reopen
		closeDB(t, db)
		mark = len(h.warnings(bucket, "w"))
		db, c = n23Open(t, crashed, bucket)
		n23CheckDec(t, "after the compaction", n23ReadDec(t, c), want)
		closeDB(t, db)
		n23Reports(t, h, "after the compaction", bucket, mark, nil)
		if n := n23Size(t, filepath.Join(crashed, bucket, "w", "head.wal")); n != 0 {
			t.Logf("head.wal holds %d bytes after the compaction", n)
		}
	})

	t.Run("two corrections for the same time", func(t *testing.T) {
		for _, damaged := range []int{0, 1} {
			t.Run(fmt.Sprintf("correction %d damaged", damaged+1), func(t *testing.T) {
				bucket := "walsame" + strconv.Itoa(damaged)
				dir := t.TempDir()
				db := openDB(t, dir, n23WALOptions())
				c := column(t, db, bucket, "w", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				for _, p := range n23WALBase() {
					must(t, c.PutI64(p.ts, p.v))
				}
				closeDB(t, db)
				db = openDB(t, dir, n23WALOptions())
				c = lookup(t, db, bucket, "w")
				path := filepath.Join(dir, bucket, "w", "head.wal")
				vals := []int64{222, 999}
				var starts []int
				for _, v := range vals {
					starts = append(starts, int(n23Size(t, path)))
					must(t, c.PutI64(n23K2, v))
				}
				starts = append(starts, int(n23Size(t, path)))
				crashed := crashCopy(t, dir)
				closeDB(t, db)
				s := n23WALSpan{path: filepath.Join(crashed, bucket, "w", "head.wal")}
				if starts[0] == 0 {
					b, err := os.ReadFile(s.path)
					must(t, err)
					_, m, _ := n23OldFirstFrame(b)
					starts[0] = m
				}
				s.edit(t, func(b []byte) []byte { return n23Flip(b, starts[damaged]+n23WALLayout(t).payloadOff) })
				want := n23WALModel(0)
				want[n23K2] = strconv.FormatInt(vals[1-damaged], 10)
				mark := len(h.warnings(bucket, "w"))
				db, c = n23Open(t, crashed, bucket)
				n23CheckDec(t, "the read", n23ReadDec(t, c), want)
				closeDB(t, db)
				n23Reports(t, h, "the open", bucket, mark, []int64{int64(starts[damaged])})
			})
		}
	})

	t.Run("DeleteRange over three points, the record of the middle one damaged", func(t *testing.T) {
		const bucket = "walrange"
		dir := t.TempDir()
		db := openDB(t, dir, n23WALOptions())
		c := column(t, db, bucket, "w", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
		for _, p := range n23WALBase() {
			must(t, c.PutI64(p.ts, p.v))
		}
		closeDB(t, db)
		db = openDB(t, dir, n23WALOptions())
		c = lookup(t, db, bucket, "w")
		must(t, c.DeleteRange(6000, 8000))
		crashed := crashCopy(t, dir)
		closeDB(t, db)
		p := filepath.Join(crashed, bucket, "w", "head.wal")
		b, err := os.ReadFile(p)
		must(t, err)
		_, m, intact := n23OldFirstFrame(b)
		if !intact || (len(b)-m)%3 != 0 {
			t.Fatalf("premise: behind the marker (%d bytes) %d bytes, not three records of one size", m, len(b)-m)
		}
		size := (len(b) - m) / 3
		mid := m + size
		must(t, os.WriteFile(p, n23Flip(b, mid+n23WALLayout(t).payloadOff), 0o644))
		want := n23WALModel(0)
		delete(want, 6000)
		delete(want, 8000)
		mark := len(h.warnings(bucket, "w"))
		db, c = n23Open(t, crashed, bucket)
		n23CheckDec(t, "the read", n23ReadDec(t, c), want)
		closeDB(t, db)
		n23Reports(t, h, "the open", bucket, mark, []int64{int64(mid)})
	})

	t.Run("a record of N23 with an intact checksum and an unknown kind", func(t *testing.T) {
		const bucket = "walunknown"
		s := n23WALNew(t, bucket)
		l := n23WALLayout(t)
		var edited []byte
		s.edit(t, func(b []byte) []byte {
			b[s.recs[1][0]+l.kindOff] = n22UnknownKind
			n23WALRenew(t, b, s.recs[1][0], s.recs[1][1])
			edited = slices.Clone(b)
			return b
		})
		db, c := n23Open(t, s.dir, bucket)
		if err := c.ScanI64(minTS, maxTS, func([]int64, []int64) bool { return true }); err == nil {
			t.Errorf("the load of the head succeeds over a record of an unknown kind with an intact checksum")
		}
		_ = c.PutI64(n23K4, 444)
		_ = db.Close()
		if now, err := os.ReadFile(s.path); err != nil || !bytes.Equal(now, edited) {
			t.Errorf("head.wal changed: something behind the record was overwritten (%v)", err)
		}
	})

	t.Run("an intact WAL, whose marker carries a version after N23", func(t *testing.T) {
		const bucket = "walversion"
		s := n23WALNew(t, bucket)
		var edited []byte
		s.edit(t, func(b []byte) []byte {
			n23WALBumpVersion(t, b)
			edited = slices.Clone(b)
			return b
		})
		db, c := n23Open(t, s.dir, bucket)
		if err := c.ScanI64(minTS, maxTS, func([]int64, []int64) bool { return true }); err == nil {
			t.Errorf("a read succeeds over a WAL of a later version")
		}
		if err := c.PutI64(n23K4, 444); err == nil {
			t.Errorf("a correction succeeds over a WAL of a later version")
		}
		_ = db.Close()
		if now, err := os.ReadFile(s.path); err != nil || !bytes.Equal(now, edited) {
			t.Errorf("head.wal changed after the attempts and Close (%v)", err)
		}
	})
}

// TestN23WALForN22: every WAL with records, which N23 creates or rewrites, read by the layout of 8ae1daaee
// (head.go:73–86), begins with an intact frame at 0 of a kind but 0 and 1, at which a stand N22 fails loudly
// (R-DEC-TS-NAGO-WAL, "Für eine Fassung N22", L-N23-32, M2). Counter row: the same read over the WAL of the image in the
// format before N23 finds an intact first frame of the kind 0 or 1.
func TestN23WALForN22(t *testing.T) {
	check := func(t *testing.T, path string) {
		t.Helper()
		b, err := os.ReadFile(path)
		must(t, err)
		kind, _, intact := n23OldFirstFrame(b)
		if !intact || kind == 0 || kind == 1 {
			t.Errorf("read by the layout of 8ae1daaee the first frame of head.wal (%d bytes) has the kind %d, intact %v: N22 would read its records", len(b), kind, intact)
		}
	}
	t.Run("a new column after a correction", func(t *testing.T) {
		s := n23WALNew(t, "forn22new")
		check(t, s.path)
	})
	t.Run("an empty WAL of testdata/f4aca80c5 after a correction", func(t *testing.T) {
		dir := n22FixtureCopy(t)
		path := filepath.Join(dir, "n22", "power", "head.wal")
		if n23Size(t, path) != 0 {
			t.Fatalf("premise: the WAL of power is not empty")
		}
		// no compaction in the background, which could fold the correction or rename files while crashCopy copies (§8)
		db := openDB(t, dir, quietOptions(0))
		c := lookup(t, db, "n22", "power")
		must(t, c.PutI64(c.Stats().MinMillis, 1))
		crashed := crashCopy(t, dir)
		closeDB(t, db)
		check(t, filepath.Join(crashed, "n22", "power", "head.wal"))
	})
	t.Run("a WAL in the format before N23 of testdata after a correction", func(t *testing.T) {
		s := n23WALOld(t, "wal", "forn22old")
		b, err := os.ReadFile(s.path)
		must(t, err)
		// counter row: the old WAL reads by the layout of 8ae1daaee as a record of the kind 0 or 1
		if kind, _, intact := n23OldFirstFrame(b); !intact || (kind != 0 && kind != 1) {
			t.Fatalf("premise: the first frame of the old WAL has the kind %d, intact %v", kind, intact)
		}
		db, c := n23Open(t, s.dir, "forn22old")
		must(t, c.PutI64(n23K4, 444))
		crashed := crashCopy(t, s.dir)
		_ = db.Close()
		check(t, filepath.Join(crashed, "forn22old", "w", "head.wal"))
	})
	t.Run("the WAL after a compaction and a correction behind it", func(t *testing.T) {
		s := n23WALNew(t, "forn22compact")
		db, c := n23Open(t, s.dir, "forn22compact")
		must(t, c.Flush())
		must(t, c.PutI64(n23K4, 444))
		crashed := crashCopy(t, s.dir)
		_ = db.Close()
		check(t, filepath.Join(crashed, "forn22compact", "w", "head.wal"))
	})
}

// TestN23OldWAL is the table "WAL im Format vor N23" of R-DEC-TS-NAGO-WAL over the images of testdata/8ae1daaee.
func TestN23OldWAL(t *testing.T) {
	h := n22cCapture(t)
	reads := n23FixtureReads(t)
	if want := n23Points(n23WALModel(4)); !slices.Equal(reads["wal"], want) || !slices.Equal(reads["walcompact"], want) {
		t.Fatalf("premise: 8ae1daaee read %v and %v from the images, the model is %v", reads["wal"], reads["walcompact"], want)
	}
	all := []int{0, 1, 2, 3}

	t.Run("an intact WAL, then a correction", func(t *testing.T) {
		const bucket = "oldwal"
		s := n23WALOld(t, "wal", bucket)
		db, c := n23Open(t, s.dir, bucket)
		want := n23WALWant(all, map[int64]int64{n23K4: 444})
		n23CheckDec(t, "before the correction", n23ReadDec(t, c), n23WALWant(all, nil))
		must(t, c.PutI64(n23K4, 444))
		n23CheckDec(t, "after the correction", n23ReadDec(t, c), want)
		crashed := crashCopy(t, s.dir)
		_ = db.Close()
		path := filepath.Join(crashed, bucket, "w", "head.wal")
		b, err := os.ReadFile(path)
		must(t, err)
		if kind, _, intact := n23OldFirstFrame(b); !intact || kind == 0 || kind == 1 {
			t.Errorf("head.wal does not begin with the marker of N23 after the correction (kind %d, intact %v)", kind, intact)
		}
		if err := checkColumnNames(filepath.Join(crashed, bucket, "w"), false); err != nil {
			t.Errorf("after the rewrite: %v", err)
		}
		db, c = n23Open(t, crashed, bucket)
		n23CheckDec(t, "after reopening", n23ReadDec(t, c), want)
		closeDB(t, db)
		if w := h.warnings(bucket, "w"); len(w) != 0 {
			t.Errorf("%d warnings, want none: %v", len(w), w)
		}
		t.Run("every record in the format of N23", func(t *testing.T) {
			if !n23WALIsNew(t, b) {
				t.Errorf("head.wal is not in the format of N23 throughout after the correction")
			}
		})
	})

	t.Run("an empty WAL of testdata/f4aca80c5, then a correction and a reopen", func(t *testing.T) {
		dir := n22FixtureCopy(t)
		path := filepath.Join(dir, "n22", "power", "head.wal")
		if n23Size(t, path) != 0 {
			t.Fatalf("premise: the WAL of power is not empty")
		}
		// no compaction in the background, which could fold the correction or rename files while crashCopy copies (§8)
		db := openDB(t, dir, quietOptions(0))
		c := lookup(t, db, "n22", "power")
		before := n23ReadDec(t, c)
		ts := c.Stats().MinMillis
		must(t, c.PutI64(ts, 4242))
		crashed := crashCopy(t, dir)
		_ = db.Close()
		b, err := os.ReadFile(filepath.Join(crashed, "n22", "power", "head.wal"))
		must(t, err)
		if kind, _, intact := n23OldFirstFrame(b); !intact || kind == 0 || kind == 1 {
			t.Errorf("the marker does not stand before the first record (kind %d, intact %v)", kind, intact)
		}
		db = openDB(t, crashed, tsdb.Options{})
		before[ts] = "4242"
		n23CheckDec(t, "after reopening", n23ReadDec(t, lookup(t, db, "n22", "power")), before)
		closeDB(t, db)
		if w := h.warnings("n22", "power"); len(w) != 0 {
			t.Errorf("%d warnings, want none: %v", len(w), w)
		}
	})

	damageK2 := func(t *testing.T, s n23WALSpan) {
		s.edit(t, func(b []byte) []byte { return n23Flip(b, s.recs[1][0]+9+1) }) // the payload of K2
	}
	unknownK2 := func(t *testing.T, s n23WALSpan) {
		s.edit(t, func(b []byte) []byte {
			b[s.recs[1][0]] = n22UnknownKind // the CRC does not cover the kind: the record stays intact by its CRC
			return b
		})
	}
	for _, d := range []struct {
		name string
		edit func(t *testing.T, s n23WALSpan)
	}{{"a byte in the payload of K2", damageK2}, {"the kind of K2 set to a value the format before N23 does not know", unknownK2}} {
		t.Run(d.name, func(t *testing.T) {
			bucket := "oldwalk2" + strconv.Itoa(len(d.name))
			s := n23WALOld(t, "wal", bucket)
			d.edit(t, s)
			db, c := n23Open(t, s.dir, bucket)
			n23CheckDec(t, "the read", n23ReadDec(t, c), n23WALWant([]int{0}, nil))
			_ = db.Close()
			n23Reports(t, h, "the open", bucket, 0, []int64{int64(s.recs[1][0])})
		})
	}

	for _, k4 := range []struct {
		name string
		v    int64
	}{{"a correction K4", 444444}, {"a correction K4 exactly as long as K2", 444}} {
		t.Run("a byte in the payload of K2, then "+k4.name+", then a reopen", func(t *testing.T) {
			bucket := "oldwalk4" + strconv.FormatInt(k4.v, 10)
			s := n23WALOld(t, "wal", bucket)
			if k4.v == 444 && len(n23OldWALFrame(0, n23K4, k4.v)) != s.recs[1][1]-s.recs[1][0] {
				t.Fatalf("premise: K4 in the format before N23 has %d bytes, K2 %d", len(n23OldWALFrame(0, n23K4, k4.v)), s.recs[1][1]-s.recs[1][0])
			}
			damageK2(t, s)
			db, c := n23Open(t, s.dir, bucket)
			must(t, c.PutI64(n23K4, k4.v))
			want := n23WALWant([]int{0}, map[int64]int64{n23K4: k4.v})
			n23CheckDec(t, "after K4", n23ReadDec(t, c), want)
			crashed := crashCopy(t, s.dir)
			_ = db.Close()
			n23Reports(t, h, "the first open", bucket, 0, []int64{int64(s.recs[1][0])})
			mark := len(h.warnings(bucket, "w"))
			db, c = n23Open(t, crashed, bucket)
			n23CheckDec(t, "after reopening", n23ReadDec(t, c), want)
			closeDB(t, db)
			n23Reports(t, h, "the reopen", bucket, mark, nil)
		})
	}

	t.Run("a byte in the payload of K2, then a compaction (Close), then a reopen", func(t *testing.T) {
		const bucket = "oldwalcompact"
		s := n23WALOld(t, "wal", bucket)
		damageK2(t, s)
		want := n23WALWant([]int{0}, nil)
		db, c := n23Open(t, s.dir, bucket)
		n23CheckDec(t, "the read", n23ReadDec(t, c), want)
		closeDB(t, db)
		n23Reports(t, h, "before the compaction", bucket, 0, []int64{int64(s.recs[1][0])})
		mark := len(h.warnings(bucket, "w"))
		db, c = n23Open(t, s.dir, bucket)
		n23CheckDec(t, "after the compaction", n23ReadDec(t, c), want)
		closeDB(t, db)
		n23Reports(t, h, "after the compaction", bucket, mark, nil)
	})

	t.Run("damaged at the first record K1, where nago tells the formats apart", func(t *testing.T) {
		// the kind, which the marker of N23 has at byte 0, read from a WAL of the code under test
		markerKind := func(t *testing.T) byte {
			s := n23WALNew(t, "oldwalmarkerkind")
			b, err := os.ReadFile(s.path)
			must(t, err)
			return b[0]
		}
		edits := []struct {
			name string
			edit func(t *testing.T, b []byte) []byte
		}{
			{"a byte in the payload of K1", func(t *testing.T, b []byte) []byte { return n23Flip(b, 9+1) }},
			{"the kind of K1 set to a value the format before N23 does not know", func(t *testing.T, b []byte) []byte { b[0] = n22UnknownKind; return b }},
			{"the kind of K1 set to the one of the marker of N23", func(t *testing.T, b []byte) []byte { b[0] = markerKind(t); return b }},
		}
		for i, e := range edits {
			t.Run(e.name, func(t *testing.T) {
				bucket := "oldwalk1" + strconv.Itoa(i)
				s := n23WALOld(t, "wal", bucket)
				s.edit(t, func(b []byte) []byte { return e.edit(t, b) })
				// K1, K2 and K3 read the values of the chunk, L is back, K4 holds
				want := n23WALWant(nil, map[int64]int64{n23K4: 444})
				db, c := n23Open(t, s.dir, bucket)
				must(t, c.PutI64(n23K4, 444))
				n23CheckDec(t, "after K4", n23ReadDec(t, c), want)
				crashed := crashCopy(t, s.dir)
				_ = db.Close()
				n23Reports(t, h, "the first open", bucket, 0, []int64{0})
				b, err := os.ReadFile(filepath.Join(crashed, bucket, "w", "head.wal"))
				must(t, err)
				if kind, _, intact := n23OldFirstFrame(b); !intact || kind == 0 || kind == 1 {
					t.Errorf("head.wal does not begin with the marker after K4 (kind %d, intact %v)", kind, intact)
				}
				mark := len(h.warnings(bucket, "w"))
				db, c = n23Open(t, crashed, bucket)
				n23CheckDec(t, "after reopening", n23ReadDec(t, c), want)
				closeDB(t, db)
				n23Reports(t, h, "the reopen", bucket, mark, nil)
			})
		}
	})

	t.Run("an image after a compaction of 8ae1daaee, which committed and was not applied", func(t *testing.T) {
		const bucket = "oldwalcommitted"
		dir := filepath.Join(t.TempDir(), "db")
		must(t, os.CopyFS(filepath.Join(dir, bucket, "w"), os.DirFS(filepath.Join(n23FixtureRoot, "db", "n23", "walcompact"))))
		db, c := n23Open(t, dir, bucket)
		n23CheckDec(t, "after the open", n23ReadDec(t, c), n23WALWant(all, nil))
		must(t, c.PutI64(n23K4, 444))
		want := n23WALWant(all, map[int64]int64{n23K4: 444})
		n23CheckDec(t, "after the correction", n23ReadDec(t, c), want)
		crashed := crashCopy(t, dir)
		_ = db.Close()
		db, c = n23Open(t, crashed, bucket)
		n23CheckDec(t, "after reopening", n23ReadDec(t, c), want)
		closeDB(t, db)
		if w := h.warnings(bucket, "w"); len(w) != 0 {
			t.Errorf("%d warnings, want none: %v", len(w), w)
		}
	})

	t.Run("a crash during the rewrite, which a correction K4 sets off", func(t *testing.T) {
		const bucket = "oldwalcrash"
		names := n23WALRewriteFiles(t)
		// the new WAL as the rewrite leaves it, with K4 behind it, and without K4: K5 behind it shows the size of a
		// correction of this size
		s := n23WALOld(t, "wal", bucket)
		colDir := filepath.Join(s.dir, bucket, "w")
		old := readSnapshot(t, colDir)
		db, c := n23Open(t, s.dir, bucket)
		must(t, c.PutI64(n23K4, 444))
		withK4, err := os.ReadFile(s.path)
		must(t, err)
		must(t, c.PutI64(n23K4+1, 444))
		k := n23Size(t, s.path) - int64(len(withK4))
		closeDB(t, db)
		rewritten := withK4[:int64(len(withK4))-k]
		var states []struct {
			files snapshot
			k4    bool
		}
		for _, name := range names {
			for n := 0; n <= len(rewritten); n++ {
				if testing.Short() && n != 0 && n != len(rewritten)/2 && n != len(rewritten) {
					continue
				}
				st := cloneSnapshot(old)
				st[name] = rewritten[:n]
				states = append(states, struct {
					files snapshot
					k4    bool
				}{st, false})
			}
		}
		for _, walIs := range []struct {
			b  []byte
			k4 bool
		}{{old["head.wal"], false}, {rewritten, false}, {withK4, true}} {
			// the states between two steps of the rewrite and of the correction: head.wal the old WAL, the new one or
			// the new one with K4, without a file of the rewrite
			st := cloneSnapshot(old)
			st["head.wal"] = walIs.b
			states = append(states, struct {
				files snapshot
				k4    bool
			}{st, walIs.k4})
		}
		for i, st := range states {
			dir := filepath.Join(t.TempDir(), "db")
			cd := filepath.Join(dir, bucket, "w")
			must(t, os.MkdirAll(cd, 0o755))
			for name, b := range st.files {
				must(t, os.WriteFile(filepath.Join(cd, name), b, 0o644))
			}
			mark := len(h.warnings(bucket, "w"))
			db, c := n23Open(t, dir, bucket)
			want := n23WALWant(all, nil)
			if st.k4 {
				want[n23K4] = "444"
			}
			n23CheckDec(t, fmt.Sprintf("state %d", i), n23ReadDec(t, c), want)
			if err := checkColumnNames(cd, false); err != nil {
				t.Errorf("state %d after the open: %v", i, err)
			}
			must(t, c.PutI64(n23K4+2, 1))
			b, err := os.ReadFile(filepath.Join(cd, "head.wal"))
			must(t, err)
			if kind, _, intact := n23OldFirstFrame(b); !intact || kind == 0 || kind == 1 {
				t.Errorf("state %d: head.wal does not begin with the marker after the next correction", i)
			}
			_ = db.Close()
			n23Reports(t, h, fmt.Sprintf("state %d", i), bucket, mark, nil)
		}
	})

	t.Run("the rewrite is kept from writing, then a correction K4", func(t *testing.T) {
		const bucket = "oldwalblocked"
		names := n23WALRewriteFiles(t)
		s := n23WALOld(t, "wal", bucket)
		colDir := filepath.Join(s.dir, bucket, "w")
		old, err := os.ReadFile(s.path)
		must(t, err)
		blocker := filepath.Join(colDir, names[0])
		must(t, os.Mkdir(blocker, 0o755))
		must(t, os.WriteFile(filepath.Join(blocker, "x"), []byte("x"), 0o644))
		db, c := n23Open(t, s.dir, bucket)
		n23CheckDec(t, "a read before", n23ReadDec(t, c), n23WALWant(all, nil))
		if err := c.PutI64(n23K4, 444); err == nil {
			t.Errorf("the correction, which sets off the rewrite, succeeds, although the rewrite cannot write %s", names[0])
		}
		if now, err := os.ReadFile(s.path); err != nil || !bytes.Equal(now, old) {
			t.Errorf("head.wal changed although the rewrite failed (%v)", err)
		}
		// counter row: without the obstacle the next correction succeeds
		must(t, os.RemoveAll(blocker))
		must(t, c.PutI64(n23K4, 444))
		want := n23WALWant(all, map[int64]int64{n23K4: 444})
		n23CheckDec(t, "after the correction", n23ReadDec(t, c), want)
		crashed := crashCopy(t, s.dir)
		_ = db.Close()
		db, c = n23Open(t, crashed, bucket)
		n23CheckDec(t, "after reopening", n23ReadDec(t, c), want)
		closeDB(t, db)
	})

	t.Run("an intact WAL, open, read and Close without a record before: the compaction of Close sets off the rewrite", func(t *testing.T) {
		img, err := os.ReadFile(filepath.Join(n23FixtureRoot, "db", "n23", "wal", "head.wal"))
		must(t, err)
		var payloads int64
		starts, _ := n22WALRecords(img)
		for _, at := range starts {
			payloads += int64(binary.LittleEndian.Uint32(img[at+1:]))
		}
		closeWrites := func(t *testing.T, bucket string, rewriteBefore bool) (int64, map[int64]string) {
			s := n23WALOld(t, "wal", bucket)
			dir := s.dir
			if rewriteBefore {
				// a correction, which repeats K1 with its value, rewrites the WAL; a crash copy keeps it unfolded
				db, c := n23Open(t, dir, bucket)
				must(t, c.PutI64(n23K1, 111))
				dir = crashCopy(t, dir)
				_ = db.Close()
			}
			pool := ndb.NewFilePool(64)
			opts := n23WALOptions()
			opts.FilePool = pool
			db, err := tsdb.Open(dir, opts)
			must(t, err)
			c := lookup(t, db, bucket, "w")
			n23CheckDec(t, "before Close", n23ReadDec(t, c), n23WALWant(all, nil))
			w0 := pool.Stats().WrittenBytes
			must(t, db.Close())
			w := pool.Stats().WrittenBytes - w0
			db = openDB(t, dir, n23WALOptions())
			got := n23ReadDec(t, lookup(t, db, bucket, "w"))
			closeDB(t, db)
			if n := n23Size(t, filepath.Join(dir, bucket, "w", "head.wal")); n != 0 {
				t.Errorf("head.wal holds %d bytes after Close: the compaction did not fold K1 to L into the chunks", n)
			}
			return w, got
		}
		wOld, gotOld := closeWrites(t, "oldwalclose", false)
		wNew, gotNew := closeWrites(t, "oldwalclosecounter", true)
		n23CheckDec(t, "the image after Close", gotOld, n23WALWant(all, nil))
		n23CheckDec(t, "counter row after Close", gotNew, gotOld)
		if wOld < wNew+payloads {
			t.Errorf("Close wrote %d bytes over the WAL in the format before N23 and %d over the rewritten one: not at least the payloads of K1 to L (%d) more, the WAL was not rewritten before the fold", wOld, wNew, payloads)
		}
		for _, b := range []string{"oldwalclose", "oldwalclosecounter"} {
			if w := h.warnings(b, "w"); len(w) != 0 {
				t.Errorf("%s: %d warnings, want none: %v", b, len(w), w)
			}
		}
	})
}
