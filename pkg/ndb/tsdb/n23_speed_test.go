// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The acceptance tests of the rows of N23 of R-DEC-TS-NAGO-SPEED (wokoda N23, L-N23-24): every chunk, which nago
// creates from N23 on, carries version 3 and keeps the prefix of version 2 with its CRC (chunk.go:51–53 of 8ae1daaee),
// so that a stand N22 refuses it as an unknown version; nago reads versions 1, 2 and 3, continues an open chunk in its
// version, removes an open chunk of version 3 shorter than its header as a torn header, and fails at a version above 3.
// The row of a record of the WAL with an unknown kind and an intact checksum is in TestN23OldWAL (format before N23: no
// error, reported) and TestN23WALDamage (format of N23: the load fails).

// n23ChunkVersions reads the version and the CRC of the prefix of every chunk of the column directory with os.
func n23ChunkVersions(t *testing.T, dir string) map[string]byte {
	t.Helper()
	m, _ := filepath.Glob(filepath.Join(dir, "*.tsb"))
	if len(m) == 0 {
		t.Fatalf("premise: no chunk in %s", dir)
	}
	vs := map[string]byte{}
	for _, p := range m {
		b, err := os.ReadFile(p)
		must(t, err)
		if !n23ChunkPrefixOK(b) {
			t.Errorf("%s: no intact prefix of version 2 (\"NAGO\" \"NTSB\", the CRC of bytes 0 … 16 at 17)", filepath.Base(p))
			continue
		}
		vs[filepath.Base(p)] = b[8]
	}
	return vs
}

// TestN23ChunkVersion3: each chunk, which nago creates from N23 on, carries version 3 with an intact prefix of version
// 2: of a new column, a new chunk after the split policy, after Flush and after a compaction. Counter row: a chunk of
// version 2 of 8ae1daaee (testdata), then more points, is read and continued in its version.
func TestN23ChunkVersion3(t *testing.T) {
	schema := tsdb.Schema{Scheme: tsdb.SchemeDecimal}
	put := func(t *testing.T, c *tsdb.Column, from, n int64) {
		for ts := from; ts < from+n; ts++ {
			must(t, c.PutI64(ts*1000, ts))
		}
	}
	all3 := func(t *testing.T, when, dir string, want int) {
		t.Helper()
		vs := n23ChunkVersions(t, dir)
		if len(vs) < want {
			t.Errorf("%s: %d chunks %v, want at least %d", when, len(vs), vs, want)
		}
		for name, v := range vs {
			if v != 3 {
				t.Errorf("%s: the chunk %s has version %d, want 3", when, name, v)
			}
		}
	}
	t.Run("a new column", func(t *testing.T) {
		dir := t.TempDir()
		db := openDB(t, dir, tsdb.Options{})
		put(t, column(t, db, "v3", "c", schema), 1, 10)
		closeDB(t, db)
		all3(t, "a new column", filepath.Join(dir, "v3", "c"), 1)
	})
	t.Run("a new chunk after the split policy", func(t *testing.T) {
		dir := t.TempDir()
		db := openDB(t, dir, tsdb.Options{BlockPoints: 2, Split: tsdb.SplitByCount(4)})
		put(t, column(t, db, "v3", "c", schema), 1, 12)
		closeDB(t, db)
		all3(t, "after the split policy", filepath.Join(dir, "v3", "c"), 2)
	})
	t.Run("after Flush", func(t *testing.T) {
		dir := t.TempDir()
		db := openDB(t, dir, tsdb.Options{})
		c := column(t, db, "v3", "c", schema)
		put(t, c, 1, 10)
		must(t, c.Flush())
		put(t, c, 20, 10)
		closeDB(t, db)
		all3(t, "after Flush", filepath.Join(dir, "v3", "c"), 2)
	})
	t.Run("after a compaction of a chunk of version 2", func(t *testing.T) {
		s := n23WALOld(t, "wal", "v3compact")
		cd := filepath.Join(s.dir, "v3compact", "w")
		if vs := n23ChunkVersions(t, cd); vs["1000_.tsb"] != 2 {
			t.Fatalf("premise: the open chunk of the image has the version %d, want 2", vs["1000_.tsb"])
		}
		db, c := n23Open(t, s.dir, "v3compact")
		must(t, c.Flush()) // folds K1 to L: the compaction writes the chunk anew
		closeDB(t, db)
		all3(t, "after the compaction", cd, 1)
		db, c = n23Open(t, s.dir, "v3compact")
		n23CheckDec(t, "after the compaction", n23ReadDec(t, c), n23WALModel(4))
		closeDB(t, db)
	})
	t.Run("counter row: a chunk of version 2 of 8ae1daaee is read and continued", func(t *testing.T) {
		s := n23WALOld(t, "wal", "v2continued")
		db, c := n23Open(t, s.dir, "v2continued")
		want := n23WALModel(4)
		for ts := int64(17000); ts <= 20000; ts += 1000 {
			must(t, c.PutI64(ts, ts))
			want[ts] = strconv.FormatInt(ts, 10)
		}
		must(t, c.Sync())
		n23CheckDec(t, "continued", n23ReadDec(t, c), want)
		crashed := crashCopy(t, s.dir)
		_ = db.Close()
		if vs := n23ChunkVersions(t, filepath.Join(crashed, "v2continued", "w")); vs["1000_.tsb"] != 2 {
			t.Errorf("the open chunk of version 2 has the version %d after it was continued, want 2: %v", vs["1000_.tsb"], vs)
		}
		db, c = n23Open(t, crashed, "v2continued")
		n23CheckDec(t, "after a crash", n23ReadDec(t, c), want)
		closeDB(t, db)
	})
}

// TestN23VersionAbove3: a chunk of version 4 with an intact prefix (set with os, its CRC renewed) of a chunk, which the
// code under test wrote, fails Column or LookupColumn (an open chunk, L-N22-27) or a scan (a finalized chunk) and stays
// byte for byte. Counter row: the same chunk with version 3 reads.
func TestN23VersionAbove3(t *testing.T) {
	for _, open := range []bool{true, false} {
		for _, v := range []byte{3, 4} {
			name := map[bool]string{true: "open chunk", false: "finalized chunk"}[open]
			t.Run(name+"/version "+strconv.Itoa(int(v)), func(t *testing.T) {
				dir := t.TempDir()
				db := openDB(t, dir, tsdb.Options{})
				c := column(t, db, "v4", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				for ts := int64(1); ts <= 20; ts++ {
					must(t, c.PutI64(ts*1000, ts))
				}
				if !open {
					must(t, c.Flush())
				}
				closeDB(t, db)
				cd := filepath.Join(dir, "v4", "c")
				glob := "*_?*.tsb"
				if open {
					glob = "*_.tsb"
				}
				m, _ := filepath.Glob(filepath.Join(cd, glob))
				if len(m) != 1 {
					t.Fatalf("premise: %d chunks %s", len(m), glob)
				}
				b, err := os.ReadFile(m[0])
				must(t, err)
				if b[8] != 3 || !n23ChunkPrefixOK(b) {
					t.Fatalf("premise: the chunk has the version %d, intact prefix %v", b[8], n23ChunkPrefixOK(b))
				}
				n23SetChunkVersion(b, v)
				must(t, os.WriteFile(m[0], b, 0o644))
				before := readSnapshot(t, cd)
				db = openDB(t, dir, tsdb.Options{})
				var errs []error
				col, ok, lerr := db.LookupColumn("v4", "c")
				if lerr == nil && !ok {
					lerr = errors.New("the column is not found")
				}
				errs = append(errs, lerr)
				_, cerr := db.Column("v4", "c", tsdb.Schema{Scheme: tsdb.SchemeDecimal})
				errs = append(errs, cerr)
				n := 0
				if lerr == nil {
					errs = append(errs, col.ScanI64(minTS, maxTS, func(ts, _ []int64) bool { n += len(ts); return true }))
				}
				_ = db.Close()
				failed := slices.ContainsFunc(errs, func(e error) bool { return e != nil })
				if v == 3 {
					if failed || n != 20 {
						t.Errorf("version 3: %v, %d points, want all 20", errs, n)
					}
					return
				}
				if !failed {
					t.Errorf("version 4: Column, LookupColumn and a scan succeed (%d points)", n)
				}
				if open && (lerr == nil || cerr == nil) {
					t.Errorf("version 4 of the open chunk: LookupColumn %v, Column %v, want both to fail", lerr, cerr)
				}
				if d := n22SnapshotDiff(before, readSnapshot(t, cd)); len(d) > 0 {
					t.Errorf("version 4: the files of the column changed: %v", d)
				}
			})
		}
	}
}

// TestN23ShortOpenChunk (R-DEC-TS-NAGO-SPEED, L-N23-32, M3): an open chunk of version 3, shorter than its header (30
// bytes with the marker and the version as in TestN22ShortOpenChunk), beside the chunks of a column, is removed as a
// torn header by Column and LookupColumn, the column opens and a read yields all its points. Counter row: the same with
// version 4, both fail and the files of the column stay byte for byte (L-N22-27).
func TestN23ShortOpenChunk(t *testing.T) {
	copyOf, srcs := n22Sources(t)
	src := srcs["fixture/multi"]
	for _, r := range []struct {
		version byte
		fails   bool
	}{{3, false}, {4, true}} {
		for _, op := range n22OpenOps {
			t.Run("version "+strconv.Itoa(int(r.version))+"/"+op, func(t *testing.T) {
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
						t.Errorf("the torn header %s of version 3 stays", name)
					}
					n := 0
					must(t, c.ScanI64(minTS, maxTS, func(ts, _ []int64) bool { n += len(ts); return true }))
					if n != src.points {
						t.Errorf("a scan sees %d points, want %d", n, src.points)
					}
					return
				}
				if err == nil {
					t.Errorf("%s succeeds over an open chunk of %d bytes of version 4", op, len(short))
				}
				if d := n22SnapshotDiff(before, after); len(d) > 0 {
					t.Errorf("%s changed the files of the column: %v", op, d)
				}
			})
		}
	}
}
