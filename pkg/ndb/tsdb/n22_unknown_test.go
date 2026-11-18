// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The acceptance tests of the loud abort (wokoda N22, L-N22-24, L-N22-27): if the marker and the format match and a
// chunk, a block, a continuation record or a record of the WAL is intact by its CRC, but its version, a flag or its kind
// is unknown, or schema.json carries an unknown version, every method with an error result (SeriesColumns included)
// fails and deletes, cuts or compacts nothing of it; a wrong marker is skipped and a torn last record of the WAL cut off
// as before. An open chunk of the right marker and an unknown version, shorter than a header, is no torn header. A Put,
// whose compaction at the cap of the head fails, succeeds, and the next Flush or Close reports the error.
//
// The carriers are copies of columns of the fixtures, edited by the format of the README and of the headers of
// block.go, chunk.go and head.go of f4aca80c5: the version is byte 8 of a chunk, whose header CRC (IEEE over bytes 0 to
// 16, at byte 17) is renewed; a flag is byte 20 of the region of a block, a kind byte 0 of the region of a record, whose
// CRC (IEEE over the region, behind it) is renewed; the kind of a record of the WAL is its first byte, which its CRC
// (over the payload) does not cover. Neither byte is masked, so the mask of a version 2 chunk stays valid without the
// names inside tsdb; the version 1 chunk of testdata/cd602526b carries an unknown flag in an unmasked block besides.
// Assumed unknown: version 0xFE, flag bit 0x80, the kind 0x7E of a record and of a record of the WAL and schema version
// 99; a stand, which gives one of them a meaning, fails the premise of the edit or must take another.
//
// Iter* drops errors (L-N22-24: "Iter* verwerfen Fehler und bleiben eine Grenze"); these tests do not ask it to abort.

const (
	n22UnknownVersion byte  = 0xFE
	n22UnknownFlag    byte  = 0x80
	n22UnknownKind    byte  = 0x7E
	n22UnknownSchema  uint8 = 99
)

// n22Unit is the edited part of a carrier: a file of the column and its bytes from off to end.
type n22Unit struct {
	file     string
	off, end int
	bytes    []byte
	prefix   []byte // the bytes of the file from its start to end
}

func n22ReadAt(t testing.TB, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	must(t, err)
	return b
}

func n22Unitof(path string, b []byte, off, end int) n22Unit {
	return n22Unit{file: filepath.Base(path), off: off, end: end, bytes: slices.Clone(b[off:end]), prefix: slices.Clone(b[:end])}
}

// n22EditVersion sets the version of the version 2 chunk at path and renews the CRC of its header.
func n22EditVersion(t testing.TB, path string, v byte) n22Unit {
	t.Helper()
	b := n22ReadAt(t, path)
	if len(b) < 85 || b[8] != 2 || crc32.ChecksumIEEE(b[:17]) != binary.LittleEndian.Uint32(b[17:]) {
		t.Fatalf("premise: %s is no intact chunk of version 2", path)
	}
	b[8] = v
	binary.LittleEndian.PutUint32(b[17:], crc32.ChecksumIEEE(b[:17]))
	must(t, os.WriteFile(path, b, 0o644))
	return n22Unitof(path, b, 0, 85)
}

// n22EditMagic breaks the marker "NAGO" of the chunk at path: a wrong marker.
func n22EditMagic(t testing.TB, path string) n22Unit {
	t.Helper()
	b := n22ReadAt(t, path)
	b[0] = 'X'
	must(t, os.WriteFile(path, b, 0o644))
	return n22Unitof(path, b, 0, 9)
}

// n22Frame returns the offset and the end of the frame of kind ('b' or 'r') with the index i, from the end if i < 0.
func n22Frame(t testing.TB, b []byte, kind byte, i int) (at, regionEnd, end int) {
	t.Helper()
	_, offs, kinds := n22ChunkFrames(t, b)
	var of []int
	for j, k := range kinds {
		if k == kind {
			of = append(of, offs[j])
		}
	}
	if i < 0 {
		i += len(of)
	}
	if i < 0 || i >= len(of) {
		t.Fatalf("premise: the chunk holds %d frames of kind %c, the test edits number %d", len(of), kind, i)
	}
	at = of[i]
	regionEnd = at + 12 + int(binary.LittleEndian.Uint32(b[at+8:]))
	return at, regionEnd, regionEnd + 4
}

// n22RenewFrame checks the CRC of the region of the frame at at before edit changes it and renews it after.
func n22RenewFrame(t testing.TB, b []byte, at, regionEnd int, edit func(region []byte)) {
	t.Helper()
	region := b[at+12 : regionEnd]
	if crc32.ChecksumIEEE(region) != binary.LittleEndian.Uint32(b[regionEnd:]) {
		t.Fatalf("premise: the frame at %d is not intact", at)
	}
	edit(region)
	binary.LittleEndian.PutUint32(b[regionEnd:], crc32.ChecksumIEEE(region))
}

// n22EditBlockFlag sets flag in the block with the index i of the chunk at path and renews its CRC.
func n22EditBlockFlag(t testing.TB, path string, i int, flag byte) n22Unit {
	t.Helper()
	b := n22ReadAt(t, path)
	at, regionEnd, end := n22Frame(t, b, 'b', i)
	n22RenewFrame(t, b, at, regionEnd, func(region []byte) {
		if region[20]&flag != 0 {
			t.Fatalf("premise: the block at %d has the flag %#x already: %#x", at, flag, region[20])
		}
		region[20] |= flag
	})
	must(t, os.WriteFile(path, b, 0o644))
	return n22Unitof(path, b, at, end)
}

// n22EditRecordKind sets the kind of the continuation record with the index i of the chunk at path, keeping the bit of
// the mask, and renews its CRC.
func n22EditRecordKind(t testing.TB, path string, i int, kind byte) n22Unit {
	t.Helper()
	b := n22ReadAt(t, path)
	at, regionEnd, end := n22Frame(t, b, 'r', i)
	n22RenewFrame(t, b, at, regionEnd, func(region []byte) {
		if k := region[0] &^ 0x80; k != 0 && k != 1 {
			t.Fatalf("premise: the record at %d has the kind %#x, not 0 or 1", at, k)
		}
		region[0] = region[0]&0x80 | kind
	})
	must(t, os.WriteFile(path, b, 0o644))
	return n22Unitof(path, b, at, end)
}

// n22EditSlotBeforeLastBlock lets the newer slot of the version 2 chunk at path end before its last block, as after
// a crash, which came after the block was written and before the slot was: a slot is | Seq 4 | End 8 | Points 8 |
// Max 8 | CRC 4 | at byte 21 or 53 (chunk.go). The open of the column then reads the last block.
func n22EditSlotBeforeLastBlock(t testing.TB, path string) {
	t.Helper()
	b := n22ReadAt(t, path)
	type slot struct {
		seq              uint32
		end, points, max int64
	}
	read := func(off int) (slot, bool) {
		s := b[off : off+32]
		if crc32.ChecksumIEEE(s[:28]) != binary.LittleEndian.Uint32(s[28:]) {
			return slot{}, false
		}
		return slot{binary.LittleEndian.Uint32(s), int64(binary.LittleEndian.Uint64(s[4:])), int64(binary.LittleEndian.Uint64(s[12:])), int64(binary.LittleEndian.Uint64(s[20:]))}, true
	}
	a, aok := read(21)
	c, cok := read(53)
	if !aok && !cok {
		t.Fatalf("premise: %s has no intact slot", path)
	}
	if !aok || cok && c.seq > a.seq {
		a = c
	}
	last, _, lastEnd := n22Frame(t, b, 'b', -1)
	prev, _, _ := n22Frame(t, b, 'b', -2)
	if a.end != int64(lastEnd) {
		t.Fatalf("premise: the newer slot ends at %d, the last block at %d", a.end, lastEnd)
	}
	lastCount := int64(binary.LittleEndian.Uint32(b[last+12+16:]))
	s := slot{seq: a.seq + 1, end: int64(last), points: a.points - lastCount, max: int64(binary.LittleEndian.Uint64(b[prev+12+8:]))}
	off := 21 + int(s.seq%2)*32
	binary.LittleEndian.PutUint32(b[off:], s.seq)
	binary.LittleEndian.PutUint64(b[off+4:], uint64(s.end))
	binary.LittleEndian.PutUint64(b[off+12:], uint64(s.points))
	binary.LittleEndian.PutUint64(b[off+20:], uint64(s.max))
	binary.LittleEndian.PutUint32(b[off+28:], crc32.ChecksumIEEE(b[off:off+28]))
	must(t, os.WriteFile(path, b, 0o644))
}

// n22EditSchemaVersion sets the version of schema.json in the column directory dir.
func n22EditSchemaVersion(t testing.TB, dir string, v uint8) n22Unit {
	t.Helper()
	path := filepath.Join(dir, "schema.json")
	var m map[string]any
	must(t, json.Unmarshal(n22ReadAt(t, path), &m))
	if m["version"] != float64(1) {
		t.Fatalf("premise: schema.json has the version %v", m["version"])
	}
	m["version"] = v
	b, err := json.Marshal(m)
	must(t, err)
	must(t, os.WriteFile(path, b, 0o644))
	return n22Unitof(path, b, 0, len(b))
}

// n22WALRecords returns the offsets of the intact records of head.wal in b and where they end: a record is | Kind 1 |
// Len 4 | CRC 4 | Payload Len |, its CRC over the payload only (head.go of f4aca80c5); the walk ends at a record, which
// is torn or whose CRC does not match.
func n22WALRecords(b []byte) (starts []int, end int) {
	for end+9 <= len(b) {
		l := int(binary.LittleEndian.Uint32(b[end+1:]))
		if end+9+l > len(b) || crc32.ChecksumIEEE(b[end+9:end+9+l]) != binary.LittleEndian.Uint32(b[end+5:]) {
			break
		}
		starts = append(starts, end)
		end += 9 + l
	}
	return starts, end
}

// n22EditWALKind sets the kind of the record with the index i of the WAL at path; the CRC does not cover it, so the
// record stays intact.
func n22EditWALKind(t testing.TB, path string, i int, kind byte) n22Unit {
	t.Helper()
	b := n22ReadAt(t, path)
	starts, end := n22WALRecords(b)
	if end != len(b) || i >= len(starts) {
		t.Fatalf("premise: the WAL holds %d intact records up to %d of %d bytes, the test edits number %d", len(starts), end, len(b), i)
	}
	at := starts[i]
	if b[at] != 0 && b[at] != 1 {
		t.Fatalf("premise: the record at %d has the kind %#x, not 0 or 1", at, b[at])
	}
	b[at] = kind
	must(t, os.WriteFile(path, b, 0o644))
	recEnd := len(b)
	if i+1 < len(starts) {
		recEnd = starts[i+1]
	}
	return n22Unitof(path, b, at, recEnd)
}

// n22Src is a database with a column to copy for a carrier.
type n22Src struct {
	name   string // the bucket/column
	schema tsdb.Schema
	points int // the points of the column
}

// n22Sources returns the databases of the carriers: a copy of the fixture of f4aca80c5 (columns multi, resume, status
// and text), of testdata/cd602526b (version 1) and a database, which the code under test wrote (fresh: a finalized
// chunk and an open chunk with a block and a continuation record).
func n22Sources(t *testing.T) (copyOf func(t *testing.T, src string) string, srcs map[string]n22Src) {
	t.Helper()
	fresh := n22FixCol{name: "fresh", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1}, compress: true}
	pts := n22Grid(n22Base, 1000, n22Values("power", 5000+4096+300))
	fresh.lives = []n22Life{{points: pts[:5000], flush: true}, {points: pts[5000:], syncAfter: []int{4096 + 300 - 1}, crash: true}}
	freshDir := n22WriteColumn(t, t.TempDir(), fresh)
	// freshwal: a WAL of the format of N23, which the code under test wrote (wokoda N23, L-N23-24): four corrections of
	// sealed points, left by a crash
	freshWAL := n22FixCol{name: "freshwal", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}, compress: true}
	walPts := n22Grid(n22Base, 1000, n22Values("power", 3000))
	var fixes []n22Point
	for _, i := range []int{10, 20, 30, 40} {
		fixes = append(fixes, n22Point{ts: walPts[i].ts, v: walPts[i].v + 1})
	}
	freshWAL.lives = []n22Life{{points: walPts}, {fix: fixes, crash: true}}
	freshWALDir := n22WriteColumn(t, t.TempDir(), freshWAL)
	dirs := map[string]string{"fixture": filepath.Join(n22FixtureRoot, "db"), "v1": filepath.Join("testdata", "cd602526b"), "fresh": freshDir,
		"freshwal": freshWALDir}
	copyOf = func(t *testing.T, src string) string {
		from, _, _ := strings.Cut(src, "/")
		name := srcs[src].name
		dir := filepath.Join(t.TempDir(), "db")
		must(t, os.CopyFS(filepath.Join(dir, name), os.DirFS(filepath.Join(dirs[from], name))))
		return dir
	}
	dec1 := tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1}
	srcs = map[string]n22Src{
		"fixture/multi":  {name: "n22/multi", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 3}, points: 10_000},
		"fixture/resume": {name: "n22/resume", schema: dec1, points: 4096 + 500 + 4096 + 300 + 200},
		"fixture/status": {name: "n22/status", schema: tsdb.Schema{Scheme: tsdb.SchemeEnum}, points: 4096 + 1000},
		"fixture/text":   {name: "n22/text", schema: tsdb.Schema{Scheme: tsdb.SchemeString}, points: 4096 + 50},
		"fixture/wal":    {name: "n22/wal", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}, points: 4096 + 100 - 2},
		"v1":             {name: "b/c", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}, points: 26},
		"fresh":          {name: "n22/fresh", schema: dec1, points: 5000 + 4096 + 300},
		"freshwal":       {name: "n22/freshwal", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}, points: 3000},
	}
	return copyOf, srcs
}

// n22Carrier is a column with one edited unit.
type n22Carrier struct {
	name string
	src  string // the key of n22Sources: "fixture/<column>", "v1" or "fresh"
	// edit edits the column directory and returns the edited unit; the counter rows run without it
	edit func(t *testing.T, colDir string) n22Unit
	// in is the glob of the chunk, into which a correction goes, at the time of its name
	in string
	// must are the operations, which meet the unit and must fail; may are the others, which may succeed, but must not
	// change the unit
	must, may []string
	// atCap are the operations, Flush or Close, which follow a Put, whose compaction at the cap of the head meets the
	// unit (TestN22CompactionInPut)
	atCap []string
	// seals are the failing operations, Flush or Close, which may seal the open block of the open chunk, whose saved
	// points are in continuation records, and finalize the chunk (decision of the Leitung on L-N22-24: appending a block
	// and finalizing delete, cut and compact nothing): for them the file of the unit keeps its bytes from its start to
	// the end of the unit under its old or a new name (n22KeptUnder), and every chunk its blocks (n22AppendedOnly),
	// instead of the files byte for byte. probe returns the unit of the column without the edit, so that the counter
	// row shows, that the check sees a compaction, which rewrites it.
	seals []string
	probe func(t *testing.T, colDir string) n22Unit
}

// n22BlockUnit is the block with the index i of the chunk at path, as it is.
func n22BlockUnit(t testing.TB, path string, i int) n22Unit {
	t.Helper()
	b := n22ReadAt(t, path)
	at, _, end := n22Frame(t, b, 'b', i)
	return n22Unitof(path, b, at, end)
}

// n22KeptUnder reports whether a chunk of the unit's time, under the old name of its file or a new one, begins with the
// bytes of the file of the unit up to the end of the unit.
func n22KeptUnder(after snapshot, u n22Unit) bool {
	lo, _, _ := strings.Cut(u.file, "_")
	for name, b := range after {
		if strings.HasPrefix(name, lo+"_") && strings.HasSuffix(name, ".tsb") && bytes.HasPrefix(b, u.prefix) {
			return true
		}
	}
	return false
}

// n22AppendedOnly names each chunk of before, which is gone or whose blocks changed in after: a chunk may have blocks
// behind its blocks, records may have given way to a block, its slots may say so and it may be finalized under a new
// name of the same start time.
func n22AppendedOnly(t testing.TB, before, after snapshot) []string {
	t.Helper()
	var out []string
	for name, b := range before {
		if !strings.HasSuffix(name, ".tsb") {
			continue
		}
		dataOff, end := n22BlocksEnd(t, b)
		lo, _, _ := strings.Cut(name, "_")
		kept := false
		for other, y := range after {
			if strings.HasPrefix(other, lo+"_") && strings.HasSuffix(other, ".tsb") && len(y) >= end && bytes.Equal(y[dataOff:end], b[dataOff:end]) {
				kept = true
			}
		}
		if !kept {
			out = append(out, name+": its blocks are gone or changed")
		}
	}
	slices.Sort(out)
	return out
}

func n22InFile(colDir, glob string) string {
	m, _ := filepath.Glob(filepath.Join(colDir, glob))
	if len(m) != 1 {
		return filepath.Join(colDir, "missing-"+glob)
	}
	return m[0]
}

var (
	n22OpenOps  = []string{"LookupColumn", "Column"}
	n22ScanI    = []string{"ScanI64", "ScanF64"}
	n22Rewrites = []string{"Close", "DeleteRange"}
	n22Writes   = []string{"Put+Sync"}
	n22AtCap    = []string{"Flush", "Close"}
)

func n22Ops(lists ...[]string) []string { return slices.Concat(lists...) }

// n22Carriers are the rows of the table: each edit beside the operations that meet it.
func n22Carriers() []n22Carrier {
	const finalMulti, openMulti = "*_?*.tsb", "*_.tsb"
	return []n22Carrier{
		{name: "unknown version of a finalized chunk", src: "fixture/multi", in: finalMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditVersion(t, n22InFile(d, "1735689600000_1735694599000.tsb"), n22UnknownVersion)
			},
			must: n22Ops(n22ScanI, n22Rewrites, []string{"Flush"}), may: n22Ops(n22OpenOps, n22Writes), atCap: n22AtCap},
		{name: "unknown flag of a block of a finalized chunk", src: "fixture/multi", in: finalMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditBlockFlag(t, n22InFile(d, "1735689600000_1735694599000.tsb"), 1, n22UnknownFlag)
			},
			must: n22Ops(n22ScanI, n22Rewrites, []string{"Flush"}), may: n22Ops(n22OpenOps, n22Writes), atCap: n22AtCap},
		{name: "unknown version of the open chunk", src: "fixture/multi", in: openMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditVersion(t, n22InFile(d, "*_.tsb"), n22UnknownVersion)
			},
			must: n22OpenOps},
		{name: "unknown flag of a sealed block of the open chunk", src: "fixture/multi", in: openMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditBlockFlag(t, n22InFile(d, "*_.tsb"), 0, n22UnknownFlag)
			},
			must: n22Ops(n22ScanI, n22Rewrites), may: n22Ops(n22OpenOps, n22Writes), atCap: []string{"Close"}},
		{name: "unknown flag of a block of the open chunk behind its slot", src: "fixture/multi", in: openMulti,
			edit: func(t *testing.T, d string) n22Unit {
				p := n22InFile(d, "*_.tsb")
				n22EditSlotBeforeLastBlock(t, p)
				return n22EditBlockFlag(t, p, -1, n22UnknownFlag)
			},
			must: n22OpenOps},
		{name: "unknown kind of a continuation record", src: "fixture/resume", in: openMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditRecordKind(t, n22InFile(d, "*_.tsb"), 0, n22UnknownKind)
			},
			must: n22OpenOps},
		{name: "unknown flag of a block of an enum column", src: "fixture/status", in: openMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditBlockFlag(t, n22InFile(d, "*_.tsb"), 0, n22UnknownFlag)
			},
			must: n22Ops([]string{"ScanString"}, n22Rewrites), may: []string{"LookupColumn", "Put+Sync"}},
		{name: "unknown flag of a block of a string column", src: "fixture/text", in: openMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditBlockFlag(t, n22InFile(d, "*_.tsb"), 0, n22UnknownFlag)
			},
			must: n22Ops([]string{"ScanString"}, n22Rewrites), may: []string{"LookupColumn", "Put+Sync"}},
		{name: "unknown version of schema.json", src: "fixture/multi", in: finalMulti,
			edit: func(t *testing.T, d string) n22Unit { return n22EditSchemaVersion(t, d, n22UnknownSchema) },
			must: n22Ops(n22OpenOps, []string{"SeriesColumns"})},
		// wokoda N23 (L-N23-24): a record of a WAL in the format before N23 with an unknown kind is a damaged record,
		// which is reported and ends the load (TestN23OldWAL); the loud abort stays for an intact record of the format of
		// N23, whose checksum covers its kind (R-DEC-TS-NAGO-WAL, L-N22-27)
		{name: "unknown kind of a record of the WAL", src: "freshwal", in: openMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n23EditWALKind(t, filepath.Join(d, "head.wal"), 1, n22UnknownKind)
			},
			must: n22Ops(n22ScanI, []string{"Put+Sync", "Put old", "DeleteRange", "Flush"}), may: n22OpenOps},
		{name: "unknown flag of an unmasked block of a version 1 chunk", src: "v1", in: "0_90.tsb",
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditBlockFlag(t, filepath.Join(d, "0_90.tsb"), 0, n22UnknownFlag)
			},
			must: n22Ops(n22ScanI, n22Rewrites, []string{"Flush"}), may: n22Ops([]string{"LookupColumn"}, n22Writes), atCap: n22AtCap},
		{name: "unknown flag of a block, which the code under test wrote", src: "fresh", in: finalMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditBlockFlag(t, n22InFile(d, finalMulti), 0, n22UnknownFlag)
			},
			must: n22Ops(n22ScanI, n22Rewrites, []string{"Flush"}), may: n22Ops([]string{"LookupColumn"}, n22Writes), atCap: n22AtCap,
			seals: n22AtCap, probe: func(t *testing.T, d string) n22Unit { return n22BlockUnit(t, n22InFile(d, finalMulti), 0) }},
		{name: "unknown kind of a record, which the code under test wrote", src: "fresh", in: openMulti,
			edit: func(t *testing.T, d string) n22Unit {
				return n22EditRecordKind(t, n22InFile(d, "*_.tsb"), 0, n22UnknownKind)
			},
			must: n22OpenOps},
	}
}

// n22Env is an operation in progress on the column of a carrier.
type n22Env struct {
	t      *testing.T
	dir    string
	bucket string
	col    string
	src    n22Src
	putAt  int64 // the first time of the chunk of the carrier, for a correction and the start of DeleteRange
	delTo  int64 // the end of DeleteRange: the last time of a finalized chunk, 10 s after the first of an open one
	db     *tsdb.DB
	mark   snapshot // the files of the column right before the operation under test
	points int      // the points a scan saw
}

func (e *n22Env) colDir() string { return filepath.Join(e.dir, e.bucket, e.col) }

// n22NewEnv returns the operation on the column of src in the database at dir, whose corrections go into the chunk of
// the glob in.
func n22NewEnv(t *testing.T, dir string, src n22Src, in string) *n22Env {
	t.Helper()
	bucket, col, _ := strings.Cut(src.name, "/")
	e := &n22Env{t: t, dir: dir, bucket: bucket, col: col, src: src}
	lo, hi, ok := n22ChunkTimes(filepath.Base(n22InFile(e.colDir(), in)))
	if !ok {
		t.Fatalf("premise: no chunk %s in %s", in, e.colDir())
	}
	e.putAt, e.delTo = lo, hi
	return e
}

// open opens the database, with opts.
func (e *n22Env) open(opts tsdb.Options) {
	db, err := tsdb.Open(e.dir, opts)
	must(e.t, err)
	e.db = db
}

// lookup returns the column; it is a step before the operation under test, whose error ends the operation.
func (e *n22Env) lookup() (*tsdb.Column, error) {
	c, ok, err := e.db.LookupColumn(e.bucket, e.col)
	if err == nil && !ok {
		err = errors.New("the column is not found")
	}
	return c, err
}

func (e *n22Env) setMark() { e.mark = readSnapshot(e.t, e.colDir()) }

func (e *n22Env) put(c *tsdb.Column, ts int64) error {
	if e.src.schema.Scheme == tsdb.SchemeDecimal {
		return c.PutI64(ts, 7)
	}
	return c.PutString(ts, "ok")
}

// n22Run runs the operation op on the column at dir and returns its error. Each operation sets the mark right before
// the call under test.
func n22Run(e *n22Env, op string) error {
	e.open(tsdb.Options{})
	e.setMark()
	switch op {
	case "LookupColumn":
		_, err := e.lookup()
		return err
	case "Column":
		_, err := e.db.Column(e.bucket, e.col, e.src.schema)
		return err
	case "SeriesColumns":
		_, err := e.db.SeriesColumns()
		return err
	}
	c, err := e.lookup()
	if err != nil {
		return err
	}
	switch op {
	case "ScanI64":
		e.setMark()
		return c.ScanI64(minTS, maxTS, func(ts, _ []int64) bool { e.points += len(ts); return true })
	case "ScanF64":
		e.setMark()
		return c.ScanF64(minTS, maxTS, func(ts []int64, _ []float64) bool { e.points += len(ts); return true })
	case "ScanString":
		e.setMark()
		return c.ScanString(minTS, maxTS, func(ts []int64, _ []string) bool { e.points += len(ts); return true })
	case "Put+Sync":
		// a point a second behind the newest, in the same quarter, so that the split policy does not end the chunk
		ts := c.Stats().MaxMillis + 1000
		e.setMark()
		return errors.Join(e.put(c, ts), c.Sync())
	case "Put old":
		// a correction, which goes into the head
		e.setMark()
		return e.put(c, e.putAt)
	case "Flush":
		if err := e.put(c, e.putAt); err != nil {
			return err
		}
		e.setMark()
		return c.Flush()
	case "Close":
		if err := e.put(c, e.putAt); err != nil {
			return err
		}
		e.setMark()
		err := e.db.Close()
		e.db = nil
		return err
	case "DeleteRange":
		e.setMark()
		return c.DeleteRange(e.putAt, e.delTo)
	}
	e.t.Fatalf("unknown operation %s", op)
	return nil
}

// n22SnapshotDiff names the files, which differ between a and b.
func n22SnapshotDiff(a, b snapshot) []string {
	var out []string
	for name, x := range a {
		y, ok := b[name]
		switch {
		case !ok:
			out = append(out, name+" is gone")
		case !bytes.Equal(x, y):
			out = append(out, name+" changed")
		}
	}
	for name := range b {
		if _, ok := a[name]; !ok {
			out = append(out, name+" is new")
		}
	}
	slices.Sort(out)
	return out
}

// n22Changes is n22SnapshotDiff without the finalization of an open chunk by Flush: <min>_.tsb gone and
// <min>_<max>.tsb new with the same bytes. A failing operation deletes, cuts or compacts nothing (L-N22-24), it keeps
// the bytes of every file; it may have finalized the open chunk, which Flush does before it compacts.
func n22Changes(a, b snapshot) []string {
	a, b = maps.Clone(a), maps.Clone(b)
	for name, x := range a {
		base, ok := strings.CutSuffix(name, "_.tsb")
		if _, still := b[name]; !ok || still {
			continue
		}
		for other, y := range b {
			if _, old := a[other]; !old && strings.HasPrefix(other, base+"_") && strings.HasSuffix(other, ".tsb") && bytes.Equal(x, y) {
				delete(a, name)
				delete(b, other)
				break
			}
		}
	}
	return n22SnapshotDiff(a, b)
}

// TestN22UnknownAborts is the table of L-N22-24 for each carrier and each operation, beside its counter row: the same
// operation on the same column and chunk without the edit succeeds and, where it writes, changes the files, a scan sees
// every point (once per column, chunk and operation). With the edit, an operation, which meets the unit, fails and
// leaves the files of the column as they were right before the call, byte for byte (n22Changes: Flush may have
// finalized the open chunk; where the carrier allows it, n22Carrier.seals, only the blocks and the unit up to its end
// are kept); the other operations may succeed, but the unit stays as it is.
func TestN22UnknownAborts(t *testing.T) {
	copyOf, srcs := n22Sources(t)
	counterRun := map[string]bool{} // a counter row runs once per column, chunk and operation
	for _, cr := range n22Carriers() {
		src := srcs[cr.src]
		for _, op := range slices.Concat(cr.must, cr.may) {
			mustFail := slices.Contains(cr.must, op)
			chunk := map[string]string{"*_?*.tsb": "finalized chunk", "*_.tsb": "open chunk"}[cr.in]
			if chunk == "" {
				chunk = cr.in
			}
			counter := "counter row/" + cr.src + ", " + chunk + "/" + op
			env := func(t *testing.T, dir string) *n22Env { return n22NewEnv(t, dir, src, cr.in) }
			seals := slices.Contains(cr.seals, op)
			if seals {
				counter += " (probe of the unit)"
			}
			if !counterRun[counter] {
				counterRun[counter] = true
				t.Run(counter, func(t *testing.T) {
					e := env(t, copyOf(t, cr.src))
					var probe n22Unit
					if seals {
						probe = cr.probe(t, e.colDir())
					}
					after := n22CounterRow(t, e, op, src.points)
					if seals && n22KeptUnder(after, probe) {
						t.Errorf("premise: without the edit %s keeps the block of the unit; a compaction, which discards it, would not be seen", op)
					}
				})
			}
			t.Run(cr.name+"/"+op, func(t *testing.T) {
				e := env(t, copyOf(t, cr.src))
				unit := cr.edit(t, e.colDir())
				err := n22Run(e, op)
				after := readSnapshot(t, e.colDir())
				if e.db != nil {
					_ = e.db.Close()
				}
				if mustFail {
					if err == nil {
						t.Errorf("%s succeeds over the %s (%d points seen)", op, cr.name, e.points)
					}
					if seals {
						if d := n22AppendedOnly(t, e.mark, after); len(d) > 0 {
							t.Errorf("%s changed the blocks of the column: %v", op, d)
						}
						if !n22KeptUnder(after, unit) {
							t.Errorf("%s did not keep %s up to the end of the %s under its old or a new name", op, unit.file, cr.name)
						}
					} else if d := n23WithoutDict(n22Changes(e.mark, after)); len(d) > 0 {
						t.Errorf("%s changed the files of the column: %v", op, d)
					}
				} else {
					t.Logf("%s: %v", op, err)
				}
				b, ok := after[unit.file]
				if !ok {
					t.Fatalf("%s removed %s, which holds the %s", op, unit.file, cr.name)
				}
				if len(b) < unit.end || !bytes.Equal(b[unit.off:unit.end], unit.bytes) {
					t.Errorf("%s changed the %s: %s bytes %d to %d", op, cr.name, unit.file, unit.off, unit.end)
				}
			})
		}
	}
}

// n22CounterRow runs op on the column without an edit: it succeeds, a scan sees every point, and an operation, which
// writes, changes the files, so that a change with the edit would be seen.
func n22CounterRow(t *testing.T, e *n22Env, op string, points int) snapshot {
	t.Helper()
	err := n22Run(e, op)
	after := readSnapshot(t, e.colDir())
	if e.db != nil {
		_ = e.db.Close()
	}
	if err != nil {
		t.Fatalf("without the edit %s fails: %v", op, err)
	}
	switch {
	case strings.HasPrefix(op, "Scan"):
		if e.points != points {
			t.Errorf("without the edit %s sees %d points, want %d", op, e.points, points)
		}
	case op == "DeleteRange" || op == "Put old":
		if bytes.Equal(e.mark["head.wal"], after["head.wal"]) {
			t.Errorf("premise: without the edit %s writes nothing into the WAL, a change with it would not be seen", op)
		}
	case op != "LookupColumn" && op != "Column" && op != "SeriesColumns":
		// Put+Sync appends a record, the others compact: a chunk changes
		if !n22ChunksChanged(e.mark, after) {
			t.Errorf("premise: without the edit %s changes no chunk, a change with it would not be seen", op)
		}
	}
	return after
}

// n22ChunkTimes returns the times of the first and the last point of a chunk from its name, of an open chunk 10 s
// after the first.
func n22ChunkTimes(name string) (lo, hi int64, ok bool) {
	base, ok := strings.CutSuffix(name, ".tsb")
	if !ok {
		return 0, 0, false
	}
	minStr, maxStr, _ := strings.Cut(base, "_")
	lo, err := strconv.ParseInt(minStr, 10, 64)
	if err != nil {
		return 0, 0, false
	}
	if maxStr == "" {
		return lo, lo + 10_000, true
	}
	hi, err = strconv.ParseInt(maxStr, 10, 64)
	return lo, hi, err == nil
}

// TestN22WrongMarkerIsSkipped: a chunk with a wrong marker is skipped as before N22 (L-N22-24 "eine falsche Kennung
// wird übersprungen wie heute"): a finalized one by a scan, which sees the points of the others without an error, an
// open one by the open of the column too; the file stays as it is. Counter row: the same column without the edit
// shows all points.
func TestN22WrongMarkerIsSkipped(t *testing.T) {
	copyOf, srcs := n22Sources(t)
	src := srcs["fixture/multi"]
	for _, r := range []struct {
		name, file string
		edit       bool
		want       int
	}{
		{"counter row: no edit", "1735689600000_1735694599000.tsb", false, src.points},
		{"a wrong marker of a finalized chunk", "1735689600000_1735694599000.tsb", true, 5000},
		{"a wrong marker of the open chunk", "1735694600000_.tsb", true, 5000},
	} {
		t.Run(r.name, func(t *testing.T) {
			dir := copyOf(t, "fixture/multi")
			colDir := filepath.Join(dir, "n22", "multi")
			var unit n22Unit
			if r.edit {
				unit = n22EditMagic(t, filepath.Join(colDir, r.file))
			}
			db := openDB(t, dir, tsdb.Options{})
			c := lookup(t, db, "n22", "multi")
			n := 0
			must(t, c.ScanI64(minTS, maxTS, func(ts, _ []int64) bool { n += len(ts); return true }))
			if n != r.want {
				t.Errorf("a scan sees %d points, want %d", n, r.want)
			}
			if r.edit {
				if b := n22ReadAt(t, filepath.Join(colDir, unit.file)); !bytes.Equal(b[unit.off:unit.end], unit.bytes) {
					t.Errorf("the chunk with the wrong marker changed")
				}
			}
		})
	}
}

// n22Logs records what slog writes, the log of nago, while it is the default logger.
type n22Logs struct {
	mu   sync.Mutex
	recs []string
}

func (l *n22Logs) Enabled(context.Context, slog.Level) bool { return true }

func (l *n22Logs) Handle(_ context.Context, r slog.Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if r.Level >= slog.LevelInfo {
		l.recs = append(l.recs, r.Level.String()+" "+r.Message)
	}
	return nil
}

func (l *n22Logs) WithAttrs([]slog.Attr) slog.Handler { return l }
func (l *n22Logs) WithGroup(string) slog.Handler      { return l }

func (l *n22Logs) records() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.recs)
}

// n22CaptureLogs makes slog write into the returned recorder until the test ends.
func n22CaptureLogs(t *testing.T) *n22Logs {
	l := &n22Logs{}
	prev := slog.Default()
	slog.SetDefault(slog.New(l))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return l
}

// n22ChunkSnapshot is the part of a snapshot, which holds chunks.
func n22ChunkSnapshot(s snapshot) snapshot {
	c := snapshot{}
	for name, b := range s {
		if strings.HasSuffix(name, ".tsb") {
			c[name] = b
		}
	}
	return c
}

// TestN22CompactionInPut (L-N22-27): a correction into a head of at most one entry compacts within Put
// (Options.MaxHeadPoints 1). Over a chunk with an unknown version or flag Put succeeds, because its point is in the WAL;
// the compaction deletes, changes and adds no chunk and writes into the log (slog, at least Info); the next Flush or
// Close reports the error and changes no file, or, where the carrier allows it (n22Carrier.seals), only seals the open
// block and finalizes the open chunk. Counter row: without the edit Put compacts, a chunk changes, and Flush or
// Close succeed.
func TestN22CompactionInPut(t *testing.T) {
	copyOf, srcs := n22Sources(t)
	for _, cr := range n22Carriers() {
		src := srcs[cr.src]
		for _, then := range cr.atCap {
			seals := slices.Contains(cr.seals, then)
			run := func(t *testing.T, edit bool) {
				e := n22NewEnv(t, copyOf(t, cr.src), src, cr.in)
				var unit n22Unit
				if edit {
					unit = cr.edit(t, e.colDir())
				} else if seals {
					unit = cr.probe(t, e.colDir())
				}
				e.open(tsdb.Options{MaxHeadPoints: 1})
				defer func() {
					if e.db != nil {
						_ = e.db.Close()
					}
				}()
				c, err := e.lookup()
				if err != nil {
					t.Fatalf("premise: the column does not open: %v", err)
				}
				logs := n22CaptureLogs(t)
				before := readSnapshot(t, e.colDir())
				putErr := e.put(c, e.putAt)
				afterPut := readSnapshot(t, e.colDir())
				compacted := len(n22SnapshotDiff(n22ChunkSnapshot(before), n22ChunkSnapshot(afterPut))) > 0
				var thenErr error
				switch then {
				case "Flush":
					thenErr = c.Flush()
				case "Close":
					thenErr = e.db.Close()
					e.db = nil
				}
				after := readSnapshot(t, e.colDir())
				if !edit {
					if putErr != nil || thenErr != nil {
						t.Fatalf("without the edit Put fails with %v and %s with %v", putErr, then, thenErr)
					}
					if !compacted {
						t.Errorf("premise: without the edit Put at the cap of the head changes no chunk, it did not compact")
					}
					if seals && n22KeptUnder(after, unit) {
						t.Errorf("premise: without the edit the compaction keeps the block of the unit; one, which discards it, would not be seen")
					}
					return
				}
				if putErr != nil {
					t.Errorf("Put fails with %v, its point is in the WAL", putErr)
				}
				if d := n22SnapshotDiff(n22ChunkSnapshot(before), n22ChunkSnapshot(afterPut)); len(d) > 0 {
					t.Errorf("the compaction in Put changed the chunks: %v", d)
				}
				if len(logs.records()) == 0 {
					t.Errorf("the failed compaction in Put writes nothing into the log")
				}
				if thenErr == nil {
					t.Errorf("the next %s reports no error", then)
				}
				if seals {
					if d := n22AppendedOnly(t, afterPut, after); len(d) > 0 {
						t.Errorf("the next %s changed the blocks of the column: %v", then, d)
					}
					if !n22KeptUnder(after, unit) {
						t.Errorf("the next %s did not keep %s up to the end of the %s under its old or a new name", then, unit.file, cr.name)
					}
				} else if d := n22Changes(afterPut, after); len(d) > 0 {
					t.Errorf("the next %s changed the files of the column: %v", then, d)
				}
				if b, ok := after[unit.file]; !ok || len(b) < unit.end || !bytes.Equal(b[unit.off:unit.end], unit.bytes) {
					t.Errorf("the %s is gone or changed (%s bytes %d to %d)", cr.name, unit.file, unit.off, unit.end)
				}
				t.Logf("Put: %v, log: %v, %s: %v", putErr, logs.records(), then, thenErr)
			}
			t.Run(cr.name+"/counter row/Put at the head cap, then "+then, func(t *testing.T) { run(t, false) })
			t.Run(cr.name+"/Put at the head cap, then "+then, func(t *testing.T) { run(t, true) })
		}
	}
}

// TestN22WALTornTail is the counter row of the record of the WAL with an unknown kind (L-N22-27, TestN22UnknownAborts):
// a torn last record of head.wal is cut off as before, a correction takes its place, and the records before it stay.
// The last record of column wal of the fixture deletes a point, which the scan sees again once it is torn.
//
// wokoda N23 (L-N23-24, R-DEC-TS-NAGO-WAL): the WAL of the fixture is in the format before N23; the correction rewrites
// it into the format of N23 first, so that the bytes of the records before the torn one do not stay where they were.
// Their effect stays: after a crash and an open the three records before the torn one and the correction hold, the
// WAL begins with the marker of N23 (read by the layout of 8ae1daaee an intact frame of a kind but 0 and 1), and no
// warning comes.
func TestN22WALTornTail(t *testing.T) {
	h := n22cCapture(t)
	copyOf, srcs := n22Sources(t)
	dir := copyOf(t, "fixture/wal")
	path := filepath.Join(dir, "n22", "wal", "head.wal")
	b := n22ReadAt(t, path)
	starts, end := n22WALRecords(b)
	if len(starts) != 4 || end != len(b) || b[starts[3]] != 1 {
		t.Fatalf("premise: the WAL holds %d intact records up to %d of %d bytes, the last of kind %#x, want 4 and a deletion", len(starts), end, len(b), b[starts[len(starts)-1]])
	}
	must(t, os.WriteFile(path, b[:len(b)-2], 0o644))
	// no compaction in the background: after the correction the head holds a tombstone of four entries, which reaches the
	// default CompactTombstoneRatio of 0.25, and a compaction, which renames its files while crashCopy copies, would make
	// the copy depend on the scheduler (AGENTS §8); the life ends with the crash copy, not with Close
	db := openDB(t, dir, quietOptions(0))
	c := lookup(t, db, "n22", "wal")
	want, err := scanMap(decimalScheme(), c)
	must(t, err)
	if n := len(want); n != srcs["fixture/wal"].points+1 {
		t.Errorf("a scan sees %d points, want %d: all but the deletion of the torn record", n, srcs["fixture/wal"].points+1)
	}
	must(t, c.PutI64(n22Base, 77))
	want[n22Base] = "77"
	crashed := crashCopy(t, dir)
	_ = db.Close()
	now := n22ReadAt(t, filepath.Join(crashed, "n22", "wal", "head.wal"))
	if kind, _, intact := n23OldFirstFrame(now); !intact || kind == 0 || kind == 1 {
		t.Errorf("the WAL does not begin with the marker of N23 after the correction (kind %d, intact %v)", kind, intact)
	}
	db = openDB(t, crashed, tsdb.Options{})
	got, err := scanMap(decimalScheme(), lookup(t, db, "n22", "wal"))
	must(t, err)
	if !mapsEqual(got, want) {
		t.Errorf("after the correction and a crash the column reads %d points, want the %d of before with the correction", len(got), len(want))
	}
	closeDB(t, db)
	if w := h.warnings("n22", "wal"); len(w) != 0 {
		t.Errorf("%d warnings, want none: a torn end is no damage: %v", len(w), w)
	}
}

// TestN22ShortOpenChunk (L-N22-27): an open chunk with the right marker and an unknown version, shorter than a header,
// is not removed as a torn header; the open of the column fails and changes no file. Counter row: the same file of
// version 2, a header torn by a crash, is removed as before and the column opens with all its points. The file is a
// second open chunk, newer than the one of column multi of the fixture.
func TestN22ShortOpenChunk(t *testing.T) {
	copyOf, srcs := n22Sources(t)
	src := srcs["fixture/multi"]
	for _, r := range []struct {
		name    string
		version byte
		fails   bool
	}{
		{"counter row: a torn header of version 2", 2, false},
		{"an unknown version", n22UnknownVersion, true},
	} {
		for _, op := range n22OpenOps {
			t.Run(r.name+"/"+op, func(t *testing.T) {
				dir := copyOf(t, "fixture/multi")
				colDir := filepath.Join(dir, "n22", "multi")
				short := make([]byte, 30)
				copy(short, "NAGONTSB")
				short[8] = r.version
				const name = "1735699600000_.tsb"
				must(t, os.WriteFile(filepath.Join(colDir, name), short, 0o644))
				before := readSnapshot(t, colDir)
				db := openDB(t, dir, tsdb.Options{})
				var c *tsdb.Column
				var err error
				if op == "Column" {
					c, err = db.Column("n22", "multi", src.schema)
				} else {
					var ok bool
					if c, ok, err = db.LookupColumn("n22", "multi"); err == nil && !ok {
						err = errors.New("the column is not found")
					}
				}
				after := readSnapshot(t, colDir)
				if !r.fails {
					if err != nil {
						t.Fatalf("%s fails: %v", op, err)
					}
					if _, ok := after[name]; ok {
						t.Errorf("the torn header %s stays", name)
					}
					n := 0
					must(t, c.ScanI64(minTS, maxTS, func(ts, _ []int64) bool { n += len(ts); return true }))
					if n != src.points {
						t.Errorf("a scan sees %d points, want %d", n, src.points)
					}
					return
				}
				if err == nil {
					t.Errorf("%s succeeds over an open chunk of %d bytes of an unknown version", op, len(short))
				}
				if d := n22SnapshotDiff(before, after); len(d) > 0 {
					t.Errorf("%s changed the files of the column: %v", op, d)
				}
			})
		}
	}
}
