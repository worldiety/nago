// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"context"
	"log/slog"
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// Tests of L-N22-29 (wokoda R-TIMESERIES-LIMITS, answer "nago meldet es"): a block tsdb skips because of a wrong CRC or
// because it cannot decode it, on any read, also when a column opens and in a compaction, is reported once per column and
// place with slog.Warn and the attributes bucket, column, file (the name of the chunk) and offset (of the block in the
// file). Reading the same block again reports nothing until the DB is opened again. A compaction that rewrites the chunk
// without the block reports it before it goes; afterwards its points are missing without a further report.
//
// The files are damaged with os only: one byte flipped at a fraction of a chunk of random values, which lies in the body
// of a block (as TestChunkBlockCorruptionSkipped does), without rebuilding the format.

// n22cRecord is a record of slog with its attributes as text.
type n22cRecord struct {
	level slog.Level
	attrs map[string]string
}

// n22cHandler keeps every record. The test that installs it runs without t.Parallel, so no other test of this package
// replaces the default meanwhile; records of other parts of the process are told apart by bucket and column.
type n22cHandler struct {
	mu    *sync.Mutex
	all   *[]n22cRecord
	attrs []slog.Attr
	group string
}

func (h *n22cHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *n22cHandler) Handle(_ context.Context, r slog.Record) error {
	rec := n22cRecord{level: r.Level, attrs: map[string]string{}}
	for _, a := range h.attrs {
		rec.attrs[a.Key] = a.Value.Resolve().String()
	}
	r.Attrs(func(a slog.Attr) bool {
		rec.attrs[h.group+a.Key] = a.Value.Resolve().String()
		return true
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.all = append(*h.all, rec)
	return nil
}

func (h *n22cHandler) WithAttrs(as []slog.Attr) slog.Handler {
	n := *h
	n.attrs = slices.Clone(h.attrs)
	for _, a := range as {
		a.Key = h.group + a.Key
		n.attrs = append(n.attrs, a)
	}
	return &n
}

func (h *n22cHandler) WithGroup(name string) slog.Handler {
	n := *h
	n.group = h.group + name + "."
	return &n
}

// warnings are the records from Warn on with the bucket and the column.
func (h *n22cHandler) warnings(bucket, column string) []n22cRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	var res []n22cRecord
	for _, r := range *h.all {
		if r.level >= slog.LevelWarn && r.attrs["bucket"] == bucket && r.attrs["column"] == column {
			res = append(res, r)
		}
	}
	return res
}

func n22cCapture(t *testing.T) *n22cHandler {
	t.Helper()
	h := &n22cHandler{mu: &sync.Mutex{}, all: &[]n22cRecord{}}
	old := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(old) })
	return h
}

const (
	n22cBlock  = 64 // points of a block
	n22cBlocks = 5  // blocks of the column
	n22cPoints = n22cBlock * n22cBlocks
	n22cColumn = "c"
)

func n22cOptions() tsdb.Options { return tsdb.Options{BlockPoints: n22cBlock, Compress: true} }

// n22cWrite writes the points 1 … n22cPoints with random values, so that no block shrinks to a few bytes; finalize
// seals the chunk with Flush, otherwise Close leaves it open with its sealed blocks.
func n22cWrite(t *testing.T, dir, bucket string, finalize bool) {
	t.Helper()
	db := openDB(t, dir, n22cOptions())
	c, err := db.Column(bucket, n22cColumn, tsdb.Schema{Scheme: tsdb.SchemeDecimal})
	must(t, err)
	r := rand.New(rand.NewPCG(22, 29))
	for ts := int64(1); ts <= n22cPoints; ts++ {
		must(t, c.PutI64(ts, r.Int64N(1_000_000_000)))
	}
	if finalize {
		must(t, c.Flush())
	}
	closeDB(t, db)
}

// n22cChunk is the largest chunk of the column whose name has (open) or has not (finalized) the suffix of an open chunk.
func n22cChunk(t *testing.T, dir, bucket string, open bool) string {
	t.Helper()
	var path string
	var size int64
	entries, err := os.ReadDir(filepath.Join(dir, bucket, n22cColumn))
	must(t, err)
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".tsb") || strings.HasSuffix(e.Name(), "_.tsb") != open {
			continue
		}
		fi, err := e.Info()
		must(t, err)
		if fi.Size() > size {
			path, size = filepath.Join(dir, bucket, n22cColumn, e.Name()), fi.Size()
		}
	}
	// at least a byte per point: the fractions lie in the bodies of the blocks, not in a header
	if path == "" || size < n22cPoints {
		t.Fatalf("premise: no chunk (open %v) of at least %d bytes in %s: %v", open, n22cPoints, filepath.Join(dir, bucket, n22cColumn), entries)
	}
	return path
}

// n22cFlip flips one byte at each fraction of the bytes lo … hi of the file (hi 0: its end) and returns the positions.
func n22cFlip(t *testing.T, path string, lo, hi int64, fractions []float64) []int64 {
	t.Helper()
	b, err := os.ReadFile(path)
	must(t, err)
	if hi == 0 {
		hi = int64(len(b))
	}
	var at []int64
	for _, f := range fractions {
		i := lo + int64(float64(hi-lo)*f)
		b[i] ^= 0xFF
		at = append(at, i)
	}
	must(t, os.WriteFile(path, b, 0o644))
	return at
}

// n22cCrash leaves in a copy of dir the disk after a crash between the last block of the open chunk and its slot, the
// one case in which the open of a column reads a block: the column is written with all blocks but the last, copied as
// after a crash (crashCopy), written with the last block and copied again; the chunk of the second copy gets the bytes
// of the first one over their length (they differ in the slot only, nago 4b22b0f84 plus the work of L-N22-29: 11 bytes
// within the first 53). So the last block is on disk and the slot does not cover it, without rebuilding the format. It returns the crashed directory, its
// open chunk and the bytes lo … hi of the last block.
func n22cCrash(t *testing.T, bucket string) (dir, chunk string, lo, hi int64) {
	t.Helper()
	live := t.TempDir()
	db := openDB(t, live, n22cOptions())
	c, err := db.Column(bucket, n22cColumn, tsdb.Schema{Scheme: tsdb.SchemeDecimal})
	must(t, err)
	r := rand.New(rand.NewPCG(22, 29))
	var before string // the disk after all blocks but the last
	for ts := int64(1); ts <= n22cPoints; ts++ {
		must(t, c.PutI64(ts, r.Int64N(1_000_000_000)))
		if ts == n22cPoints-n22cBlock {
			before = crashCopy(t, live)
		}
	}
	dir = crashCopy(t, live)
	closeDB(t, db)
	chunk = n22cChunk(t, dir, bucket, true)
	first, err := os.ReadFile(n22cChunk(t, before, bucket, true))
	must(t, err)
	b, err := os.ReadFile(chunk)
	must(t, err)
	if len(b) <= len(first) {
		t.Fatalf("premise: the last block is not on disk after the crash (%d bytes, %d before it)", len(b), len(first))
	}
	copy(b, first) // the first bytes as before the last block (they differ in the slot only), the last block behind them
	must(t, os.WriteFile(chunk, b, 0o644))
	return dir, chunk, int64(len(first)), int64(len(b))
}

// n22cGaps are the runs of missing points of 1 … n22cPoints in a scan of the whole column.
func n22cGaps(t *testing.T, c *tsdb.Column) [][2]int64 {
	t.Helper()
	seen := map[int64]bool{}
	must(t, c.ScanI64(0, n22cPoints+1, func(ts, _ []int64) bool {
		for _, x := range ts {
			seen[x] = true
		}
		return true
	}))
	var gaps [][2]int64
	for ts := int64(1); ts <= n22cPoints; ts++ {
		if seen[ts] {
			continue
		}
		if len(gaps) > 0 && gaps[len(gaps)-1][1] == ts-1 {
			gaps[len(gaps)-1][1] = ts
		} else {
			gaps = append(gaps, [2]int64{ts, ts})
		}
	}
	return gaps
}

// TestN22CorruptBlockReported: L-N22-29 as a table. Every row with damage has its counter row with the same steps on an
// intact column, which reads whole and reports nothing.
func TestN22CorruptBlockReported(t *testing.T) {
	h := n22cCapture(t)
	ran := 0
	for i, c := range []struct {
		name      string
		open      bool      // the damaged chunk stays open (Close without Flush)
		crash     bool      // the damaged block lies behind the slot of the open chunk after a crash (n22cCrash)
		compact   bool      // a compaction rewrites the chunk before any read
		fractions []float64 // damaged places; none: the counter row
	}{
		{name: "scan over a block with a wrong CRC", fractions: []float64{0.5}},
		{name: "scan over two damaged blocks at different places", fractions: []float64{0.25, 0.75}},
		{name: "counter row: scan over an intact column"},
		{name: "compaction over a damaged block without a read before", compact: true, fractions: []float64{0.5}},
		{name: "counter row: compaction over an intact column", compact: true},
		// after a Close the open of a column reads no block of the open chunk; a scan reads and reports it
		{name: "damaged block of the open chunk after Close, reported by a scan", open: true, fractions: []float64{0.5}},
		{name: "counter row: intact open chunk after Close", open: true},
		// after a crash between a block and its slot the open reads the block behind the slot and reports it
		{name: "damaged block behind the slot after a crash, reported when the column opens", crash: true, fractions: []float64{0.5}},
		{name: "counter row: intact block behind the slot after a crash", crash: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			ran++
			bucket := "n22corrupt" + strconv.Itoa(i)
			want := len(c.fractions)
			var dir, chunk string
			var at []int64
			if c.crash {
				var lo, hi int64
				dir, chunk, lo, hi = n22cCrash(t, bucket)
				at = n22cFlip(t, chunk, lo, hi, c.fractions)
			} else {
				dir = t.TempDir()
				n22cWrite(t, dir, bucket, !c.open)
				chunk = n22cChunk(t, dir, bucket, c.open)
				at = n22cFlip(t, chunk, 0, 0, c.fractions)
			}
			before, err := os.ReadFile(chunk)
			must(t, err)

			// check: count records of this column and that each names the chunk and an offset at or before its byte
			check := func(t *testing.T, when string, n int) {
				t.Helper()
				got := h.warnings(bucket, n22cColumn)
				if len(got) != n {
					t.Errorf("%s: %d warnings with bucket %s and column %s, want %d: %v", when, len(got), bucket, n22cColumn, n, got)
					return
				}
				// each damaged byte lies in its own block: the k-th place (sorted by offset) is at or before the k-th
				// damaged byte and after the one before it; a reopen reports every place once more
				count := map[int64]int{}
				for _, r := range got {
					off, err := strconv.ParseInt(r.attrs["offset"], 10, 64)
					if filepath.Base(r.attrs["file"]) != filepath.Base(chunk) || err != nil {
						t.Errorf("%s: warning %v, want file %s and a number as offset", when, r.attrs, filepath.Base(chunk))
					}
					count[off]++
				}
				if n == 0 {
					return
				}
				offs := slices.Sorted(maps.Keys(count))
				if len(offs) != len(at) {
					t.Fatalf("%s: offsets %v, want one place for each damaged byte %v", when, count, at)
				}
				for k, off := range offs {
					if off > at[k] || (k > 0 && off <= at[k-1]) || count[off] != n/len(at) {
						t.Errorf("%s: offsets %v, want each in its own block at or before the damaged bytes %v, %d times", when, count, at, n/len(at))
					}
				}
			}
			gapsOK := func(t *testing.T, when string, gaps [][2]int64) {
				t.Helper()
				if len(gaps) != want {
					t.Errorf("%s: missing points %v, want the points of %d blocks", when, gaps, want)
				}
				for _, g := range gaps {
					if g[1]-g[0]+1 > n22cBlock {
						t.Errorf("%s: missing points %v, more than a block of %d", when, g, n22cBlock)
					}
				}
			}

			db := openDB(t, dir, n22cOptions())
			col, ok, err := db.LookupColumn(bucket, n22cColumn)
			if err != nil || !ok {
				t.Fatalf("LookupColumn: %v %v", ok, err)
			}
			switch {
			case c.open || c.crash:
				if c.crash {
					check(t, "after opening the column", want)
				} else if n := len(h.warnings(bucket, n22cColumn)); n > want {
					t.Errorf("after opening the column: %d warnings, want at most %d", n, want)
				}
				gapsOK(t, "a scan after the open", n22cGaps(t, col))
				check(t, "after a scan", want)
				closeDB(t, db)
			case c.compact:
				// an overwrite of the first point with its own value, so that the compaction of Flush rewrites the chunk
				first := rand.New(rand.NewPCG(22, 29)).Int64N(1_000_000_000) // the value n22cWrite wrote at 1
				must(t, col.PutI64(1, first))
				must(t, col.Flush())
				closeDB(t, db) // waits for the compaction
				check(t, "after the compaction", want)
				if now, err := os.ReadFile(chunk); err == nil && slices.Equal(now, before) {
					t.Fatalf("premise: the compaction did not rewrite %s", chunk)
				}
				db = openDB(t, dir, n22cOptions())
				col, ok, err = db.LookupColumn(bucket, n22cColumn)
				if err != nil || !ok {
					t.Fatalf("LookupColumn after the compaction: %v %v", ok, err)
				}
				gapsOK(t, "a scan after the compaction", n22cGaps(t, col))
				check(t, "a scan after the compaction (the block is gone)", want)
				closeDB(t, db)
			default:
				gapsOK(t, "the first scan", n22cGaps(t, col))
				check(t, "after the first scan", want)
				gapsOK(t, "the second scan", n22cGaps(t, col))
				check(t, "after the second scan", want)
				closeDB(t, db)
				db = openDB(t, dir, n22cOptions())
				col, ok, err = db.LookupColumn(bucket, n22cColumn)
				if err != nil || !ok {
					t.Fatalf("LookupColumn after reopening: %v %v", ok, err)
				}
				gapsOK(t, "a scan after reopening", n22cGaps(t, col))
				check(t, "after reopening and a scan", 2*want)
				closeDB(t, db)
			}
		})
	}
	if ran != 9 {
		t.Fatalf("premise: %d rows ran, want 9", ran)
	}
}
