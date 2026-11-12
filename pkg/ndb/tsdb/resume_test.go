// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

// The tests of this file show that a column keeps its open chunk over DB.Close and over a crash: the next open continues
// it, and only the split policy ends it. A life is one open of the database, ended by Close or by a crash.

// chunkNames are the names of the open and of the finalized chunk files of a column, sorted.
func chunkNames(t *testing.T, dir, bucket, column string) (pending, final []string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, bucket, column))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, e := range entries {
		switch name := e.Name(); {
		case strings.HasSuffix(name, "_"+chunkExt):
			pending = append(pending, name)
		case strings.HasSuffix(name, chunkExt):
			final = append(final, name)
		}
	}
	return pending, final
}

// allChunkFiles counts the chunk files below dir, of every bucket and column.
func allChunkFiles(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(d.Name(), chunkExt) {
			n++
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return fi.Size()
}

func openColumnIn(t *testing.T, dir string, opts Options, schema Schema) (*DB, *Column) {
	t.Helper()
	db, err := Open(dir, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	c, err := db.Column("b", "c", schema)
	if err != nil {
		t.Fatal(err)
	}
	return db, c
}

func closeDB(t *testing.T, db *DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestCloseContinuesTheOpenChunk: a column written in ten lives, each ended by DB.Close, has as many chunk files as the
// same points written in one life, as long as the split policy ends no chunk: one open chunk. A life writes 6 points in
// blocks of 4, so that Close finds a block on disk and points in memory. Flush before Close is what Close did before (it
// finalized the open chunk); its row shows that the files are counted and grow by one in each life.
func TestCloseContinuesTheOpenChunk(t *testing.T) {
	const lives, perLife = 10, 6
	opts := Options{BlockPoints: 4}
	for _, s := range syncSchemes() {
		// the reference: the same points in one life
		ref := t.TempDir()
		refDB, refCol := openColumnIn(t, ref, opts, s.schema)
		for i := range lives * perLife {
			if err := s.put(refCol, i); err != nil {
				t.Fatal(err)
			}
		}
		refPending, refFinal := chunkNames(t, ref, "b", "c")
		closeDB(t, refDB)
		if len(refPending) != 1 || len(refFinal) != 0 {
			t.Fatalf("%s: premise: the split policy ends no chunk of the reference: %v %v", s.name, refPending, refFinal)
		}

		for _, tt := range []struct {
			name        string
			flush       bool
			wantPending int
			wantFinal   func(lives int) int
		}{
			{name: "Close", wantPending: len(refPending), wantFinal: func(int) int { return len(refFinal) }},
			{name: "Flush before Close, as Close did before", flush: true, wantPending: 0, wantFinal: func(n int) int { return n }},
		} {
			t.Run(s.name+"/"+tt.name, func(t *testing.T) {
				dir := t.TempDir()
				var firstName string
				for life := range lives {
					db, c := openColumnIn(t, dir, opts, s.schema)
					checkSyncPoints(t, s, c, upTo(life*perLife)) // what the lives before wrote
					for i := range perLife {
						if err := s.put(c, life*perLife+i); err != nil {
							t.Fatal(err)
						}
					}
					if tt.flush {
						if err := c.Flush(); err != nil {
							t.Fatal(err)
						}
					}
					closeDB(t, db)

					pending, final := chunkNames(t, dir, "b", "c")
					if len(pending) != tt.wantPending || len(final) != tt.wantFinal(life+1) {
						t.Fatalf("after life %d: %d open %v and %d finalized files %v, want %d and %d",
							life+1, len(pending), pending, len(final), final, tt.wantPending, tt.wantFinal(life+1))
					}
					if len(pending) == 1 {
						if firstName == "" {
							firstName = pending[0]
						}
						if pending[0] != firstName {
							t.Fatalf("after life %d the open chunk is %s, not %s of the first life", life+1, pending[0], firstName)
						}
					}
				}
				checkSyncPoints(t, s, reopenColumn(t, dir, opts), upTo(lives*perLife))
			})
		}
	}
}

// TestSplitPolicyEndsAContinuedChunk: the split policy ends a chunk continued over Close as in a single life, by time and
// by size. A life is one day with a point each hour, in blocks of 16, so Close finds points in memory. The reference
// writes the same points in one life with a Sync at the end of each day, as the statistics of wokoda do. A chunk ended
// by the policy is finalized and does not change any more, also not by the lives after it.
func TestSplitPolicyEndsAContinuedChunk(t *testing.T) {
	const (
		day   = int64(24 * 60 * 60 * 1000)
		hour  = day / 24
		lives = 15
	)
	// grid is the rule of extapi.StatsOptions in wokoda: a chunk lies in one step of a grid of span
	grid := func(span int64) SplitFunc {
		return func(s ChunkStats) bool { return s.NextMillis/span != s.MinMillis/span }
	}
	const size = 420 // bytes: about four days of these points
	put := func(c *Column, d, h int64) error { return c.PutI64(d*day+h*hour, (d*24+h)*7%13) }
	for _, tt := range []struct {
		name      string
		split     SplitFunc
		byGrid    bool
		wantFinal int
	}{
		{name: "6 days as in extapi.StatsOptions", split: CombineSplits(SplitBySize(64<<20), grid(6*day)), byGrid: true, wantFinal: 2},
		{name: "size", split: SplitBySize(size), wantFinal: 3},
		{name: "size within 6 days", split: CombineSplits(SplitBySize(size), grid(6*day)), byGrid: true, wantFinal: 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{BlockPoints: 16, Compress: true, Split: tt.split}

			ref := t.TempDir()
			refDB, refCol := openColumnIn(t, ref, opts, Schema{Scheme: SchemeDecimal})
			for d := range int64(lives) {
				for h := range int64(24) {
					if err := put(refCol, d, h); err != nil {
						t.Fatal(err)
					}
				}
				if err := refCol.Sync(); err != nil {
					t.Fatal(err)
				}
			}
			refPending, refFinal := chunkNames(t, ref, "b", "c")
			closeDB(t, refDB)

			dir := t.TempDir()
			sealed := map[string][]byte{} // each finalized chunk as it was in the life that finalized it
			for d := range int64(lives) {
				db, c := openColumnIn(t, dir, opts, Schema{Scheme: SchemeDecimal})
				for h := range int64(24) {
					if err := put(c, d, h); err != nil {
						t.Fatal(err)
					}
				}
				closeDB(t, db)
				_, final := chunkNames(t, dir, "b", "c")
				for _, name := range final {
					if _, ok := sealed[name]; !ok {
						sealed[name] = readFile(t, filepath.Join(dir, "b", "c", name))
					}
				}
			}

			pending, final := chunkNames(t, dir, "b", "c")
			if !slices.Equal(pending, refPending) || !slices.Equal(final, refFinal) {
				t.Errorf("chunks %v %v, in a single life %v %v", pending, final, refPending, refFinal)
			}
			if len(final) != tt.wantFinal || len(pending) != 1 {
				t.Errorf("%d finalized and %d open chunks, want %d and 1: %v %v", len(final), len(pending), tt.wantFinal, final, pending)
			}
			continued := 0
			for _, name := range final {
				info, _ := parseChunkName(name)
				if info.maxMillis/day > info.minMillis/day {
					continued++ // it holds points of more than one life
				}
				if info.maxMillis-info.minMillis >= 6*day {
					t.Errorf("chunk %s spans 6 days or more", name)
				}
				if tt.byGrid && info.minMillis/(6*day) != info.maxMillis/(6*day) {
					t.Errorf("chunk %s lies in two steps of 6 days", name)
				}
				if !bytes.Equal(readFile(t, filepath.Join(dir, "b", "c", name)), sealed[name]) {
					t.Errorf("chunk %s changed after it was finalized", name)
				}
			}
			if continued != len(final) {
				t.Errorf("premise: only %d of the finalized chunks %v were continued after a Close", continued, final)
			}

			c := reopenColumn(t, dir, opts)
			ts, vals := collectI64(t, c, minInt64, maxInt64)
			if len(ts) != lives*24 {
				t.Fatalf("%d points, want %d", len(ts), lives*24)
			}
			for i := range ts {
				d, h := int64(i)/24, int64(i)%24
				if ts[i] != d*day+h*hour || vals[i] != (d*24+h)*7%13 {
					t.Fatalf("point %d is %d=%d", i, ts[i], vals[i])
				}
			}
		})
	}
}

// TestCrashContinuesTheOpenChunk: a crash leaves the open chunk without Close. The next open takes it up to its last
// intact block, cuts off what a torn write left behind it and continues it: the same file, no finalized one. Three
// Syncs write blocks of 2, 3 and 1 points.
func TestCrashContinuesTheOpenChunk(t *testing.T) {
	const blocks = 3
	perBlock := [blocks]int{2, 3, 1}
	for _, tt := range []struct {
		name       string
		tear       func(t *testing.T, path string, sizes [blocks]int64)
		wantBlocks int // the intact blocks after the crash
	}{
		{name: "nothing torn", tear: func(*testing.T, string, [blocks]int64) {}, wantBlocks: 3},
		{name: "the last block cut short", wantBlocks: 2, tear: func(t *testing.T, path string, sizes [blocks]int64) {
			if err := os.Truncate(path, sizes[2]-3); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "the last block garbled", wantBlocks: 2, tear: func(t *testing.T, path string, sizes [blocks]int64) {
			b := readFile(t, path)
			b[sizes[2]-6] ^= 0xFF // in its body: the length fits, the CRC does not
			if err := os.WriteFile(path, b, 0644); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "the beginning of a block behind the last one", wantBlocks: 3, tear: func(t *testing.T, path string, sizes [blocks]int64) {
			b := readFile(t, path)
			torn := append(b, b[sizes[1]:sizes[1]+20]...) // sync marker, length and bounds of a block without its body
			if err := os.WriteFile(path, torn, 0644); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "two blocks torn", wantBlocks: 1, tear: func(t *testing.T, path string, sizes [blocks]int64) {
			if err := os.Truncate(path, sizes[1]-1); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{BlockPoints: 4}
			s := syncSchemes()[0]
			dir := t.TempDir()
			_, c := openColumnIn(t, dir, opts, s.schema)
			var sizes [blocks]int64
			n := 0
			for b := range blocks {
				for range perBlock[b] {
					if err := s.put(c, n); err != nil {
						t.Fatal(err)
					}
					n++
				}
				if err := c.Sync(); err != nil {
					t.Fatal(err)
				}
				pending, _ := syncChunkFiles(t, dir)
				sizes[b] = fileSize(t, pending[0])
			}

			crashed := crashCopy(t, dir)
			pending, final := chunkNames(t, crashed, "b", "c")
			if len(pending) != 1 || len(final) != 0 {
				t.Fatalf("premise: the crash leaves one open chunk: %v %v", pending, final)
			}
			path := filepath.Join(crashed, "b", "c", pending[0])
			tt.tear(t, path, sizes)

			want := 0
			for b := range tt.wantBlocks {
				want += perBlock[b]
			}
			c2 := reopenColumn(t, crashed, opts)
			checkSyncPoints(t, s, c2, upTo(want))
			if got := fileSize(t, path); got != sizes[tt.wantBlocks-1] {
				t.Errorf("the open chunk has %d bytes after the open, it had %d after the last intact Sync", got, sizes[tt.wantBlocks-1])
			}
			if p, f := chunkNames(t, crashed, "b", "c"); !slices.Equal(p, pending) || len(f) != 0 {
				t.Errorf("after the open: open chunks %v and finalized %v, want only %v", p, f, pending)
			}

			// the next life goes on in the same file, and a second crash keeps both lives
			for i := range 2 {
				if err := s.put(c2, n+i); err != nil {
					t.Fatal(err)
				}
			}
			if err := c2.Sync(); err != nil {
				t.Fatal(err)
			}
			wantAll := append(upTo(want), n, n+1)
			again := crashCopy(t, crashed)
			checkSyncPoints(t, s, reopenColumn(t, again, opts), wantAll)
			if p, f := chunkNames(t, again, "b", "c"); !slices.Equal(p, pending) || len(f) != 0 {
				t.Errorf("after the second crash: open chunks %v and finalized %v, want only %v", p, f, pending)
			}
		})
	}
}

// TestOpenContinuesChunksOfAnEarlierVersion opens what nago cd602526b, before this change, left on disk in
// testdata/cd602526b: a decimal column b/c in blocks of 4 with the points 0, 10 ... 250 of the value time+1. Its first
// two lives ended with Close, which finalized 0_90.tsb and 100_190.tsb; the third life wrote 200..250 and a Sync and then
// crashed, which left the open chunk 200_.tsb. Nothing is converted: the finalized chunks stay byte for byte, and the
// open one is continued, also over a Close.
func TestOpenContinuesChunksOfAnEarlierVersion(t *testing.T) {
	opts := Options{BlockPoints: 4}
	fixture := filepath.Join("testdata", "cd602526b")
	dir := filepath.Join(t.TempDir(), "db")
	if err := os.CopyFS(dir, os.DirFS(fixture)); err != nil {
		t.Fatal(err)
	}
	wantFinal := []string{"0_90.tsb", "100_190.tsb"}
	wantPending := []string{"200_.tsb"}
	if p, f := chunkNames(t, dir, "b", "c"); !slices.Equal(p, wantPending) || !slices.Equal(f, wantFinal) {
		t.Fatalf("premise: the fixture holds %v and %v", p, f)
	}
	original := map[string][]byte{}
	for _, name := range wantFinal {
		original[name] = readFile(t, filepath.Join(dir, "b", "c", name))
	}

	db, err := Open(dir, opts)
	if err != nil {
		t.Fatal(err)
	}
	c, ok, err := db.LookupColumn("b", "c")
	if err != nil || !ok {
		t.Fatalf("lookup: %v %v", ok, err)
	}
	checkPoints(t, c, every10(0, 26))
	if st := c.Stats(); st.MinMillis != 0 || st.MaxMillis != 250 || st.Chunks != 3 {
		t.Errorf("stats: %+v", st)
	}
	putEvery10(t, c, 26, 6) // a block and points in memory
	checkPoints(t, c, every10(0, 32))
	closeDB(t, db)

	for life := range 2 {
		if p, f := chunkNames(t, dir, "b", "c"); !slices.Equal(p, wantPending) || !slices.Equal(f, wantFinal) {
			t.Fatalf("after life %d: %v and %v, want %v and %v", life+1, p, f, wantPending, wantFinal)
		}
		for name, b := range original {
			if !bytes.Equal(readFile(t, filepath.Join(dir, "b", "c", name)), b) {
				t.Errorf("after life %d: the finalized chunk %s changed", life+1, name)
			}
		}
		db, err := Open(dir, opts)
		if err != nil {
			t.Fatal(err)
		}
		c, ok, err := db.LookupColumn("b", "c")
		if err != nil || !ok {
			t.Fatalf("lookup: %v %v", ok, err)
		}
		checkPoints(t, c, every10(0, 32+life))
		putEvery10(t, c, 32+life, 1)
		if n, err := c.Count(); err != nil || n != int64(33+life) {
			t.Errorf("count: %d %v", n, err)
		}
		closeDB(t, db)
	}
}

// TestCloseFoldsTheHeadAndKeepsTheChunkOpen: Close folds the head into the chunks, so that the next life reads in order.
// A change before the open chunk rewrites only finalized chunks, a change in it rewrites the open chunk, which stays
// open. Chunks end after 8 points in blocks of 4: 0..70 is finalized, 80..110 is open. Flush instead of Close finalizes,
// as before.
func TestCloseFoldsTheHeadAndKeepsTheChunkOpen(t *testing.T) {
	opts := Options{BlockPoints: 4, Split: SplitByCount(8)}
	const finalized, open = "0_70.tsb", "80_.tsb"
	for _, tt := range []struct {
		name           string
		change         func(c *Column) error
		flush          bool
		want           map[int64]int64 // the changes of the values of time+1, 0 for a deleted point
		finalChanged   bool
		openChanged    bool
		wantOpen       []string
		wantFinalAfter int // finalized chunks after the next life
	}{
		{name: "no change", change: func(*Column) error { return nil }, wantOpen: []string{open}, wantFinalAfter: 1},
		{name: "a correction in a finalized chunk", change: func(c *Column) error { return c.PutI64(30, 999) },
			want: map[int64]int64{30: 999}, finalChanged: true, wantOpen: []string{open}, wantFinalAfter: 1},
		{name: "a correction in the open chunk", change: func(c *Column) error { return c.PutI64(90, 999) },
			want: map[int64]int64{90: 999}, openChanged: true, wantOpen: []string{open}, wantFinalAfter: 1},
		{name: "a new point in the open chunk", change: func(c *Column) error { return c.PutI64(95, 96) },
			want: map[int64]int64{95: 96}, openChanged: true, wantOpen: []string{open}, wantFinalAfter: 1},
		{name: "a deletion in the open chunk", change: func(c *Column) error { return c.Delete(100) },
			want: map[int64]int64{100: 0}, openChanged: true, wantOpen: []string{open}, wantFinalAfter: 1},
		{name: "a change in both", change: func(c *Column) error { return errors.Join(c.PutI64(30, 999), c.PutI64(90, 999)) },
			want: map[int64]int64{30: 999, 90: 999}, finalChanged: true, openChanged: true, wantOpen: []string{open}, wantFinalAfter: 1},
		{name: "the open chunk deleted", change: func(c *Column) error { return c.DeleteRange(80, 110) },
			want: map[int64]int64{80: 0, 90: 0, 100: 0, 110: 0}, openChanged: true, wantOpen: nil, wantFinalAfter: 1},
		{name: "Flush before Close finalizes, as Close did before", change: func(c *Column) error { return c.PutI64(90, 999) }, flush: true,
			want: map[int64]int64{90: 999}, openChanged: true, wantOpen: nil, wantFinalAfter: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db, c := openColumnIn(t, dir, opts, Schema{Scheme: SchemeDecimal})
			putEvery10(t, c, 0, 12)
			if p, f := chunkNames(t, dir, "b", "c"); !slices.Equal(p, []string{open}) || !slices.Equal(f, []string{finalized}) {
				t.Fatalf("premise: %v %v", p, f)
			}
			finalBefore := readFile(t, filepath.Join(dir, "b", "c", finalized))
			openBefore := readFile(t, filepath.Join(dir, "b", "c", open))
			if err := tt.change(c); err != nil {
				t.Fatal(err)
			}
			if tt.flush {
				if err := c.Flush(); err != nil {
					t.Fatal(err)
				}
			}
			closeDB(t, db)

			if c.head.len() != 0 {
				t.Errorf("Close left %d entries in the head", c.head.len())
			}
			pending, final := chunkNames(t, dir, "b", "c")
			if !slices.Equal(pending, tt.wantOpen) {
				t.Errorf("open chunks %v, want %v", pending, tt.wantOpen)
			}
			if !tt.flush {
				if !slices.Equal(final, []string{finalized}) {
					t.Errorf("finalized chunks %v, want %v", final, finalized)
				} else if changed := !bytes.Equal(readFile(t, filepath.Join(dir, "b", "c", finalized)), finalBefore); changed != tt.finalChanged {
					t.Errorf("the finalized chunk changed: %v, want %v", changed, tt.finalChanged)
				}
				if len(pending) == 1 {
					if changed := !bytes.Equal(readFile(t, filepath.Join(dir, "b", "c", open)), openBefore); changed != tt.openChanged {
						t.Errorf("the open chunk changed: %v, want %v", changed, tt.openChanged)
					}
				}
			}

			want := map[int64]int64{}
			for _, ts := range every10(0, 14) {
				want[ts] = ts + 1
			}
			for ts, v := range tt.want {
				want[ts] = v
				if v == 0 {
					delete(want, ts)
				}
			}
			wantTs := make([]int64, 0, len(want))
			for ts := range want {
				wantTs = append(wantTs, ts)
			}
			slices.Sort(wantTs)

			// the next life reads in order and continues the open chunk
			db2, c2 := openColumnIn(t, dir, opts, Schema{Scheme: SchemeDecimal})
			putEvery10(t, c2, 12, 2)
			closeDB(t, db2)
			for life, c := range []*Column{c2, reopenColumn(t, dir, opts)} {
				ts, vals := collectI64(t, c, minInt64, maxInt64)
				if !slices.Equal(ts, wantTs) {
					t.Fatalf("life %d reads %v, want %v", life+2, ts, wantTs)
				}
				for i := range ts {
					if vals[i] != want[ts[i]] {
						t.Fatalf("life %d: the value of %d is %d, want %d", life+2, ts[i], vals[i], want[ts[i]])
					}
				}
			}
			pending, final = chunkNames(t, dir, "b", "c")
			if len(pending) != 1 || len(final) != tt.wantFinalAfter {
				t.Errorf("after the next life: open %v and finalized %v, want one open and %d finalized", pending, final, tt.wantFinalAfter)
			}
		})
	}
}

// TestDropBeforeAContinuedChunk: retention removes a continued open chunk only as a whole, when all its points are older
// than the limit, those of the earlier life, those on disk and those in memory alike. Chunks end after 8 points in
// blocks of 4: the first life writes 0..110, so 0..70 is finalized and 80..110 open; the second life may add 120 and 130
// in memory. The state after it survives more points and a restart.
func TestDropBeforeAContinuedChunk(t *testing.T) {
	for _, tt := range []struct {
		name     string
		newer    int // points of the second life
		cut      int64
		removed  int
		from     int // index of the first remaining point
		files    int
		wantOpen bool
	}{
		{name: "before all points", newer: 2, cut: 0, removed: 0, from: 0, files: 2, wantOpen: true},
		{name: "after the finalized chunk", newer: 2, cut: 71, removed: 1, from: 8, files: 1, wantOpen: true},
		{name: "at the newest point of the earlier life", newer: 2, cut: 110, removed: 1, from: 8, files: 1, wantOpen: true},
		{name: "after the earlier life, before the points in memory", newer: 2, cut: 111, removed: 1, from: 8, files: 1, wantOpen: true},
		{name: "at the newest point in memory", newer: 2, cut: 130, removed: 1, from: 8, files: 1, wantOpen: true},
		{name: "after all points", newer: 2, cut: 131, removed: 2, from: 14, files: 0},
		{name: "only the earlier life: at its newest point", newer: 0, cut: 110, removed: 1, from: 8, files: 1, wantOpen: true},
		{name: "only the earlier life: after its newest point", newer: 0, cut: 111, removed: 2, from: 12, files: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db := openDropDB(t, dir)
			putEvery10(t, decimalColumn(t, db, "b", "c"), 0, 12)
			closeDB(t, db)

			db = openDropDB(t, dir)
			c := decimalColumn(t, db, "b", "c")
			putEvery10(t, c, 12, tt.newer)
			total := 12 + tt.newer
			removed, err := c.DropBefore(tt.cut)
			if err != nil {
				t.Fatal(err)
			}
			if removed != tt.removed {
				t.Errorf("removed %d, want %d", removed, tt.removed)
			}
			want := every10(tt.from, total)
			checkPoints(t, c, want)
			pending, final := chunkNames(t, dir, "b", "c")
			if len(pending)+len(final) != tt.files || (len(pending) == 1) != tt.wantOpen {
				t.Errorf("files: open %v, finalized %v, want %d files, open %v", pending, final, tt.files, tt.wantOpen)
			}
			if st := c.Stats(); st.Chunks != tt.files {
				t.Errorf("stats: %+v, want %d chunks", st, tt.files)
			}

			putEvery10(t, c, 100, 3)
			want = append(want, every10(100, 103)...)
			checkPoints(t, c, want)
			closeDB(t, db)
			checkPoints(t, decimalColumn(t, openDropDB(t, dir), "b", "c"), want)
		})
	}
}

// TestDropAContinuedChunk drops a column and a bucket, whose open chunks the second life continues, next to columns that
// keep theirs. The dropped name starts empty in a new file.
func TestDropAContinuedChunk(t *testing.T) {
	for _, tt := range []struct {
		name string
		drop func(db *DB) error
		keep []string // bucket/column that survive
	}{
		{name: "column", drop: func(db *DB) error { return db.DropColumn("b", "c") }, keep: []string{"b/keep", "other/c"}},
		{name: "bucket", drop: func(db *DB) error { return db.DropBucket("b") }, keep: []string{"other/c"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			opts := Options{BlockPoints: 4}
			all := []string{"b/c", "b/keep", "other/c"}
			column := func(db *DB, name string) *Column {
				bucket, col, _ := strings.Cut(name, "/")
				return decimalColumn(t, db, bucket, col)
			}
			db, err := Open(dir, opts)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range all {
				putEvery10(t, column(db, name), 0, 6)
			}
			closeDB(t, db)

			db, err = Open(dir, opts)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			cols := map[string]*Column{}
			for _, name := range all {
				cols[name] = column(db, name)
				putEvery10(t, cols[name], 6, 3)
			}
			if err := tt.drop(db); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(dir, "b", "c")); !os.IsNotExist(err) {
				t.Errorf("the column is still on disk: %v", err)
			}
			if err := cols["b/c"].PutI64(1000, 1); !errors.Is(err, errColumnGone) {
				t.Errorf("an old handle wrote: %v", err)
			}
			fresh := column(db, "b/c")
			checkPoints(t, fresh, nil)
			putEvery10(t, fresh, 50, 1)
			for _, name := range tt.keep {
				checkPoints(t, cols[name], every10(0, 9))
			}
			closeDB(t, db)

			db, err = Open(dir, opts)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			checkPoints(t, column(db, "b/c"), every10(50, 51))
			for _, name := range tt.keep {
				checkPoints(t, column(db, name), every10(0, 9))
			}
			for _, name := range append([]string{"b/c"}, tt.keep...) {
				bucket, col, _ := strings.Cut(name, "/")
				if p, f := chunkNames(t, dir, bucket, col); len(p) != 1 || len(f) != 0 {
					t.Errorf("%s: open %v, finalized %v, want one open chunk", name, p, f)
				}
			}
		})
	}
}

// TestStatsCountTheOpenChunk: the statistics of a column include its open chunk, which lives on over a restart: Chunks
// and Bytes are the chunk files on disk, and Count finds every point, also those only in memory.
func TestStatsCountTheOpenChunk(t *testing.T) {
	const points = 10 // in blocks of 4: two on disk, two points in memory
	opts := Options{BlockPoints: 4}
	for _, tt := range []struct {
		name  string
		after func(t *testing.T, dir string, db *DB, c *Column) *Column
	}{
		{name: "running", after: func(_ *testing.T, _ string, _ *DB, c *Column) *Column { return c }},
		{name: "after Sync", after: func(t *testing.T, _ string, _ *DB, c *Column) *Column {
			if err := c.Sync(); err != nil {
				t.Fatal(err)
			}
			return c
		}},
		{name: "after Close", after: func(t *testing.T, dir string, db *DB, _ *Column) *Column {
			closeDB(t, db)
			return reopenColumn(t, dir, opts)
		}},
		{name: "after Flush", after: func(t *testing.T, _ string, _ *DB, c *Column) *Column {
			if err := c.Flush(); err != nil {
				t.Fatal(err)
			}
			return c
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db, c := openColumnIn(t, dir, opts, Schema{Scheme: SchemeDecimal})
			putEvery10(t, c, 0, points)
			c = tt.after(t, dir, db, c)

			pending, final := chunkNames(t, dir, "b", "c")
			var files int64
			for _, name := range append(pending, final...) {
				files += fileSize(t, filepath.Join(dir, "b", "c", name))
			}
			st := c.Stats()
			if !st.HasData || st.MinMillis != 0 || st.MaxMillis != (points-1)*10 {
				t.Errorf("bounds: %+v", st)
			}
			if st.Chunks != len(pending)+len(final) || st.Bytes != files {
				t.Errorf("stats %d chunks of %d bytes, on disk %d files of %d bytes", st.Chunks, st.Bytes, len(pending)+len(final), files)
			}
			if n, err := c.Count(); err != nil || n != points {
				t.Errorf("count %d %v, want %d", n, err, points)
			}
		})
	}
}

// TestRestartsKeepOneFilePerColumn is the case of the measurement M18 of wokoda in small: 50 columns of one bucket, which
// a process writes in ten lives, each ended by Close. The threshold, set before the run: at most one chunk file per
// column, i.e. 50. Flush before Close, which is what Close did before, gives one file per column and life, 500.
func TestRestartsKeepOneFilePerColumn(t *testing.T) {
	const columns, lives, perLife = 50, 10, 20
	for _, tt := range []struct {
		name     string
		flush    bool
		maxFiles int
		minFiles int
	}{
		{name: "Close", maxFiles: columns, minFiles: columns},
		{name: "Flush before Close, as Close did before", flush: true, maxFiles: columns * lives, minFiles: columns * lives},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for life := range lives {
				db, err := Open(dir, Options{})
				if err != nil {
					t.Fatal(err)
				}
				for col := range columns {
					c := decimalColumn(t, db, "area", fmt.Sprint("series", col))
					putEvery10(t, c, life*perLife, perLife)
					if tt.flush {
						if err := c.Flush(); err != nil {
							t.Fatal(err)
						}
					}
				}
				closeDB(t, db)
			}
			files := allChunkFiles(t, dir)
			t.Logf("%d columns, %d lives: %d chunk files (threshold %d)", columns, lives, files, tt.maxFiles)
			if files > tt.maxFiles || files < tt.minFiles {
				t.Errorf("%d chunk files, want %d to %d", files, tt.minFiles, tt.maxFiles)
			}

			db, err := Open(dir, Options{})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			for col := range columns {
				checkPoints(t, decimalColumn(t, db, "area", fmt.Sprint("series", col)), every10(0, lives*perLife))
			}
		})
	}
}

// TestCompactionKeepsTheChunkOpenWhileWriting: the compaction of the head rewrites the open chunk while it is written
// and read, and keeps it open. Run with -race. An appender writes the times 1, 2 ... in batches; a corrector overwrites
// the newest points of each batch, MaxHeadPoints of them, so that its last put compacts synchronously. The two are
// interleaved by force: batch b+1 is appended while round b is corrected. Readers see a prefix 1..m in order at any time.
// At the end the column has every point with its newest value in its one open chunk, also after a restart. There is one
// corrector only, and it is the only one to compact, because a put into the head is not fenced against a compaction
// running at the same time.
func TestCompactionKeepsTheChunkOpenWhileWriting(t *testing.T) {
	const (
		batches, perBatch = 20, 100
		readers           = 4
	)
	opts := Options{BlockPoints: 16, MaxHeadPoints: 8}
	dir := t.TempDir()
	db, c := openColumnIn(t, dir, opts, Schema{Scheme: SchemeDecimal})

	appended := make([]chan struct{}, batches)
	corrected := make([]chan struct{}, batches)
	for b := range batches {
		appended[b], corrected[b] = make(chan struct{}), make(chan struct{})
	}
	want := make(map[int64]int64, batches*perBatch)
	for ts := int64(1); ts <= batches*perBatch; ts++ {
		want[ts] = ts
	}
	for b := range batches {
		last := int64(b+1) * perBatch
		for j := range int64(opts.MaxHeadPoints) {
			want[last-j] = -(last - j)
		}
	}

	var errs [2 + readers]error
	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Go(func() { // the appender
		for b := range batches {
			if b > 1 {
				<-corrected[b-2] // round b-2 is done, round b-1 may run along with this batch
			}
			for i := range perBatch {
				ts := int64(b*perBatch + i + 1)
				if err := c.PutI64(ts, ts); err != nil {
					errs[0] = err
					return
				}
				if ts%37 == 0 {
					if err := c.Sync(); err != nil {
						errs[0] = err
						return
					}
				}
			}
			close(appended[b])
		}
	})
	wg.Go(func() { // the corrector
		for b := range batches {
			<-appended[b]
			last := int64(b+1) * perBatch
			for j := range int64(opts.MaxHeadPoints) {
				if err := c.PutI64(last-j, -(last - j)); err != nil {
					errs[1] = err
					return
				}
			}
			if n := c.head.len(); n != 0 {
				errs[1] = fmt.Errorf("round %d: the head kept %d entries, the compaction did not run", b, n)
				return
			}
			close(corrected[b])
		}
	})
	for r := range readers {
		wg.Go(func() {
			for {
				select {
				case <-done:
					return
				default:
				}
				next := int64(1)
				err := c.ScanI64(minInt64, maxInt64, func(ts, _ []int64) bool {
					for _, got := range ts {
						if got != next {
							errs[2+r] = fmt.Errorf("a reader saw %d where %d was next", got, next)
							return false
						}
						next++
					}
					return true
				})
				if err != nil || errs[2+r] != nil {
					errs[2+r] = errors.Join(errs[2+r], err)
					return
				}
			}
		})
	}
	<-corrected[batches-1]
	close(done)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}

	check := func(name string, c *Column) {
		t.Helper()
		ts, vals := collectI64(t, c, minInt64, maxInt64)
		if len(ts) != len(want) {
			t.Fatalf("%s: %d points, want %d", name, len(ts), len(want))
		}
		for i := range ts {
			if ts[i] != int64(i)+1 || vals[i] != want[ts[i]] {
				t.Fatalf("%s: point %d is %d=%d, want %d=%d", name, i, ts[i], vals[i], i+1, want[int64(i)+1])
			}
		}
		if p, f := chunkNames(t, dir, "b", "c"); len(p) != 1 || len(f) != 0 {
			t.Errorf("%s: open %v, finalized %v, want one open chunk", name, p, f)
		}
	}
	check("running", c)
	closeDB(t, db)
	check("after a restart", reopenColumn(t, dir, opts))
}
