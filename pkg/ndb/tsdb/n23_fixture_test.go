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
	"hash/crc32"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The images in the format before N23 (wokoda N23, R-DEC-TS-NAGO-DICT, R-DEC-TS-NAGO-WAL, "Die Abbilder liegen in
// testdata, geschrieben mit nago 8ae1daaee oder nach der Beschreibung des Formats von Hand gebaut"). TestN23WriteFixture
// wrote them on nago 8ae1daaee, before the work of N23; the columns are in testdata/8ae1daaee/db/n23:
//
//   - dict: an enum column with the values A, B, C and D (1 000, 1 008, 1 016 and 1 024 bytes, so that the entries of B
//     and C together are a multiple of the header of 8 bytes of enumdict.go:23–27), each with points in a sealed block
//     (1 000 … 8 000), in a continuation record (9 000 … 12 000, after a Sync) and D as a correction in the WAL (500),
//     left by a crash: enum.dict and head.wal in the format before N23, the open chunk of version 2;
//   - dicttail: dict and a fifth value X (1 032 bytes) at 13 000, which only the memory held at the crash: the last entry
//     of enum.dict carries no point;
//   - wal: a decimal column with the points 1 000 … 16 000 (value i·10) in two sealed blocks of the open chunk, then
//     out of order the corrections K1 (2 000 = 111), K2 (3 000 = 222), K3 (4 000 = 333) and the deletion L (5 000), left
//     by a crash: head.wal in the format before N23 with four records;
//   - walcompact: wal after a compaction of 8ae1daaee, which committed and was not applied: compact.log, the new chunk
//     under its name compact-n23-0.new, which folds K1 and K2, its new WAL head-n23.wal.new with K3, and the old WAL
//     with K1 to L, of which L lies behind walSize. The new files are those of a compaction of 8ae1daaee over the
//     state after K1 and K2 (Flush, Close), the manifest and the new WAL are built by the format of compact.go and
//     head.go of 8ae1daaee.
//
// expect.json holds what 8ae1daaee read from each column, each point as "<time>=<value>" (an enum value by its name A to
// D or X).
//
//	N23_WRITE_FIXTURE=1 go test -run '^TestN23WriteFixture$' ./pkg/ndb/tsdb/
//
// Without N23_WRITE_FIXTURE=1 it does nothing; on a later stand it would write that stand's files instead.

const n23FixtureRoot = "testdata/8ae1daaee"

// n23DictLens are the lengths of the texts of A, B, C, D and X.
var n23DictLens = map[string]int{"A": 1000, "B": 1008, "C": 1016, "D": 1024, "X": 1032}

// n23Text returns the text of the enum value name with n bytes: the name, a dash and random lower case letters, so
// that a byte in its middle lies in its text and no text is a part of another.
func n23Text(name string, n int) string {
	var seed uint64
	for _, c := range name {
		seed = seed*131 + uint64(c)
	}
	r := rand.New(rand.NewPCG(seed, uint64(n)))
	b := []byte(name + "-")
	for len(b) < n {
		b = append(b, byte('a'+r.IntN(26)))
	}
	return string(b[:n])
}

// n23Value is the text of one of the values A to D and X.
func n23Value(name string) string { return n23Text(name, n23DictLens[name]) }

// n23NameOf maps a text of n23Value back to its name, for the messages and expect.json; another text stays as it is.
func n23NameOf(s string) string {
	for name := range n23DictLens {
		if s == n23Value(name) {
			return name
		}
	}
	return s
}

// n23DictOptions: blocks of 8 points and no compaction in the background, so that a crash copy holds what the column
// wrote and nothing of a compaction in its middle.
func n23DictOptions() tsdb.Options { return quietOptions(8) }

// n23DictPoints are the points of A to D of column dict: in a sealed block, in a continuation record and D as a
// correction in the WAL.
func n23DictPoints() (sealed, record []n22Point, wal n22Point) {
	names := []string{"A", "B", "C", "D"}
	for i := range 8 {
		sealed = append(sealed, n22Point{ts: int64(i+1) * 1000, str: n23Value(names[i%4])})
	}
	for i := range 4 {
		record = append(record, n22Point{ts: int64(9+i) * 1000, str: n23Value(names[i])})
	}
	return sealed, record, n22Point{ts: 500, str: n23Value("D")}
}

// n23WriteDict writes column bucket/name with the points of n23DictPoints into the database at dir and returns the
// crash copy after them, and the sizes of enum.dict after each of A, B, C and D was written first. With tail, X is
// written at 13 000 before the crash and stays in memory.
func n23WriteDict(t testing.TB, dir, bucket, name string, tail bool) (crashed string, dictEnds map[string]int64) {
	t.Helper()
	db := openDB(t, dir, n23DictOptions())
	c := column(t, db, bucket, name, tsdb.Schema{Scheme: tsdb.SchemeEnum})
	dictPath := filepath.Join(dir, bucket, name, "enum.dict")
	dictEnds = map[string]int64{}
	sealed, record, wal := n23DictPoints()
	for _, p := range slices.Concat(sealed, record) {
		must(t, c.PutString(p.ts, p.str))
		if n := n23NameOf(p.str); dictEnds[n] == 0 {
			dictEnds[n] = n23Size(t, dictPath)
		}
	}
	must(t, c.Sync())
	must(t, c.PutString(wal.ts, wal.str))
	if tail {
		must(t, c.PutString(13000, n23Value("X")))
		dictEnds["X"] = n23Size(t, dictPath)
	}
	crashed = crashCopy(t, dir)
	closeDB(t, db)
	return crashed, dictEnds
}

// n23Size is the size of the file at path, 0 if there is none.
func n23Size(t testing.TB, path string) int64 {
	t.Helper()
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0
	}
	must(t, err)
	return fi.Size()
}

// n23WALOptions: blocks of 8 points and no compaction in the background.
func n23WALOptions() tsdb.Options { return quietOptions(8) }

// The corrections and the deletion of column wal and their times.
const (
	n23K1, n23K2, n23K3, n23L, n23K4 = int64(2000), int64(3000), int64(4000), int64(5000), int64(6000)
)

// n23WALBase are the points 1 000 … 16 000 of column wal with the value i·10.
func n23WALBase() []n22Point {
	var p []n22Point
	for i := 1; i <= 16; i++ {
		p = append(p, n22Point{ts: int64(i) * 1000, v: int64(i) * 10})
	}
	return p
}

// n23WALSteps are the steps of the second life of column wal: K1, K2, K3 as corrections, L as a deletion.
var n23WALSteps = []struct {
	ts  int64
	v   int64
	del bool
}{{n23K1, 111, false}, {n23K2, 222, false}, {n23K3, 333, false}, {n23L, 0, true}}

// n23WriteWAL writes column bucket/name: the points of n23WALBase and Close, then the steps of n23WALSteps up to
// stop (exclusive) and a crash copy after them, which it returns; its column holds head.wal with these records. The
// sizes of head.wal after each step are in walEnds.
func n23WriteWAL(t testing.TB, dir, bucket, name string, stop int) (crashed string, walEnds []int64) {
	t.Helper()
	db := openDB(t, dir, n23WALOptions())
	c := column(t, db, bucket, name, tsdb.Schema{Scheme: tsdb.SchemeDecimal})
	for _, p := range n23WALBase() {
		must(t, c.PutI64(p.ts, p.v))
	}
	closeDB(t, db)
	db = openDB(t, dir, n23WALOptions())
	c = lookup(t, db, bucket, name)
	walPath := filepath.Join(dir, bucket, name, "head.wal")
	for _, s := range n23WALSteps[:stop] {
		if s.del {
			must(t, c.Delete(s.ts))
		} else {
			must(t, c.PutI64(s.ts, s.v))
		}
		walEnds = append(walEnds, n23Size(t, walPath))
	}
	crashed = crashCopy(t, dir)
	closeDB(t, db)
	return crashed, walEnds
}

// n23WALModel is what column wal reads with the records of the first k steps of n23WALSteps applied.
func n23WALModel(k int) map[int64]string {
	m := map[int64]string{}
	for _, p := range n23WALBase() {
		m[p.ts] = strconv.FormatInt(p.v, 10)
	}
	for _, s := range n23WALSteps[:k] {
		if s.del {
			delete(m, s.ts)
		} else {
			m[s.ts] = strconv.FormatInt(s.v, 10)
		}
	}
	return m
}

// n23OldWALFrame is a record of the WAL in the format before N23 (head.go:73–86 of 8ae1daaee): | Kind 1 | Len 4 |
// CRC32 4 over the payload | payload |, the payload the zig-zag varints of the time and, for a put, of the value.
func n23OldWALFrame(kind byte, ts, v int64) []byte {
	payload := binary.AppendVarint(nil, ts)
	if kind == 0 {
		payload = binary.AppendVarint(payload, v)
	}
	var hdr [9]byte
	hdr[0] = kind
	binary.LittleEndian.PutUint32(hdr[1:], uint32(len(payload)))
	binary.LittleEndian.PutUint32(hdr[5:], crc32.ChecksumIEEE(payload))
	return append(hdr[:], payload...)
}

// TestN23WriteFixture writes testdata/8ae1daaee with the code of the stand it runs on (see the top of this file); it
// ran on nago 8ae1daaee, before the work of N23, and checks there what each column reads.
func TestN23WriteFixture(t *testing.T) {
	if os.Getenv("N23_WRITE_FIXTURE") != "1" {
		t.Skip("writes testdata/8ae1daaee only with N23_WRITE_FIXTURE=1")
	}
	must(t, os.RemoveAll(n23FixtureRoot))
	put := func(name, from string) {
		must(t, os.CopyFS(filepath.Join(n23FixtureRoot, "db", "n23", name), os.DirFS(from)))
	}

	// dict and dicttail
	crashed, _ := n23WriteDict(t, t.TempDir(), "n23", "dict", false)
	put("dict", filepath.Join(crashed, "n23", "dict"))
	crashed, _ = n23WriteDict(t, t.TempDir(), "n23", "dicttail", true)
	put("dicttail", filepath.Join(crashed, "n23", "dicttail"))

	// wal
	crashed, walEnds := n23WriteWAL(t, t.TempDir(), "n23", "wal", 4)
	walDir := filepath.Join(crashed, "n23", "wal")
	put("wal", walDir)

	// walcompact: the new chunks of a compaction over the state after K1 and K2
	k12, _ := n23WriteWAL(t, t.TempDir(), "n23", "wal", 2)
	db := openDB(t, k12, n23WALOptions())
	must(t, lookup(t, db, "n23", "wal").Flush())
	closeDB(t, db)
	compacted := readSnapshot(t, filepath.Join(k12, "n23", "wal"))
	img := readSnapshot(t, walDir)
	m := struct {
		Remove  []string    `json:"remove"`
		Rename  [][2]string `json:"rename"`
		WAL     string      `json:"wal"`
		WALSize int64       `json:"walSize"`
	}{WAL: "head-n23.wal.new", WALSize: walEnds[2]}
	finals := map[string]bool{}
	for _, name := range slices.Sorted(func(yield func(string) bool) {
		for n := range compacted {
			if strings.HasSuffix(n, ".tsb") && !yield(n) {
				return
			}
		}
	}) {
		tmp := fmt.Sprintf("compact-n23-%d.new", len(m.Rename))
		img[tmp] = compacted[name]
		m.Rename = append(m.Rename, [2]string{tmp, name})
		finals[name] = true
	}
	for name := range img {
		if strings.HasSuffix(name, ".tsb") && !finals[name] {
			m.Remove = append(m.Remove, name)
		}
	}
	slices.Sort(m.Remove)
	if len(m.Rename) == 0 || len(m.Remove) == 0 {
		t.Fatalf("premise: the compaction wrote %v and replaces %v", m.Rename, m.Remove)
	}
	img["head-n23.wal.new"] = n23OldWALFrame(0, n23K3, 333)
	if !bytes.Equal(img["head.wal"][walEnds[1]:walEnds[2]], img["head-n23.wal.new"]) {
		t.Fatalf("premise: the record of K3 in the old WAL is not the one built by the format of 8ae1daaee")
	}
	b, err := json.Marshal(m)
	must(t, err)
	img["compact.log"] = b
	cdir := filepath.Join(n23FixtureRoot, "db", "n23", "walcompact")
	must(t, os.MkdirAll(cdir, 0o755))
	for name, b := range img {
		must(t, os.WriteFile(filepath.Join(cdir, name), b, 0o644))
	}

	// what this stand reads from the images, opened from a copy, so that the open does not change them
	reads := map[string][]string{}
	db = openDB(t, n23FixtureCopy(t), tsdb.Options{})
	for _, col := range []string{"dict", "dicttail", "wal", "walcompact"} {
		s := decimalScheme()
		if strings.HasPrefix(col, "dict") {
			s = enumScheme()
		}
		got, err := scanMap(s, lookup(t, db, "n23", col))
		must(t, err)
		reads[col] = n23Points(got)
	}
	closeDB(t, db)
	if want := n23Points(n23WALModel(4)); !slices.Equal(reads["wal"], want) || !slices.Equal(reads["walcompact"], want) {
		t.Errorf("wal reads %v and walcompact %v, the model is %v", reads["wal"], reads["walcompact"], want)
	}
	if want := n23Points(n23DictModel()); !slices.Equal(reads["dict"], want) || !slices.Equal(reads["dicttail"], want) {
		t.Errorf("dict reads %v and dicttail %v, the model is %v", reads["dict"], reads["dicttail"], want)
	}
	n22WriteJSON(t, filepath.Join(n23FixtureRoot, "expect.json"), reads)
}

// n23DictModel is what column dict reads: the points of n23DictPoints by the names of their values.
func n23DictModel() map[int64]string {
	sealed, record, wal := n23DictPoints()
	m := map[int64]string{}
	for _, p := range slices.Concat(sealed, record, []n22Point{wal}) {
		m[p.ts] = n23NameOf(p.str)
	}
	return m
}

// n23Points lists the points of m as "<time>=<value>" in time order, an enum value by its name.
func n23Points(m map[int64]string) []string {
	var out []string
	for _, ts := range sortedKeys(m) {
		out = append(out, fmt.Sprintf("%d=%s", ts, n23NameOf(m[ts])))
	}
	return out
}

// n23FixtureCopy copies the database of testdata/8ae1daaee to a new directory and returns it.
func n23FixtureCopy(t testing.TB) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "db")
	must(t, os.CopyFS(dir, os.DirFS(filepath.Join(n23FixtureRoot, "db"))))
	return dir
}

// n23FixtureReads returns what 8ae1daaee read from each column of testdata/8ae1daaee.
func n23FixtureReads(t testing.TB) map[string][]string {
	t.Helper()
	var reads map[string][]string
	n22ReadJSON(t, filepath.Join(n23FixtureRoot, "expect.json"), &reads)
	if len(reads) != 4 {
		t.Fatalf("premise: expect.json holds %d columns, want 4", len(reads))
	}
	return reads
}
