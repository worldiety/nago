// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// fillCols creates n decimal columns each pre-filled with `fill` points, flushed
// to sealed chunks.
func fillCols(tb testing.TB, db *DB, prefix string, n int, fill int64) []*Column {
	tb.Helper()
	cols := make([]*Column, n)
	base := int64(1_700_000_000_000)
	for i := 0; i < n; i++ {
		c, err := db.Column("b", fmt.Sprintf("%s%d", prefix, i), Schema{Scheme: SchemeDecimal, Decimals: 2})
		if err != nil {
			tb.Fatal(err)
		}
		for j := int64(0); j < fill; j++ {
			c.PutI64(base+j*20, 2000+(j%13))
		}
		c.Flush()
		cols[i] = c
	}
	return cols
}

// TestConcurrentColumnsScaleAndCorrect verifies that operations on different
// columns are independent: they run without data races (run with -race), do not
// serialize on a shared column lock, and aggregate read throughput rises with
// the number of concurrent columns. Reads and writes only take the per-column
// lock and the sharded FilePool, so disjoint columns proceed in parallel.
func TestConcurrentColumnsScaleAndCorrect(t *testing.T) {
	const fill = 1_000_000
	const workers = 8
	const repeats = 6 // repeated scans per worker to average out setup/GC noise

	db, _ := Open(t.TempDir(), Options{Compress: false})
	defer db.Close()
	cols := fillCols(t, db, "c", workers, fill)

	// single-worker baseline: one column scanned `repeats` times.
	singleStart := time.Now()
	scanAll(t, cols[0], repeats)
	singleRate := float64(fill*repeats) / time.Since(singleStart).Seconds()

	// all workers scan their own column concurrently, `repeats` times each.
	allStart := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(c *Column) {
			defer wg.Done()
			scanAll(t, c, repeats)
		}(cols[i])
	}
	wg.Wait()
	allRate := float64(int64(fill)*repeats*workers) / time.Since(allStart).Seconds()

	speedup := allRate / singleRate
	t.Logf("read scaling: 1 worker %.0f M/s, %d workers %.0f M/s -> %.2fx",
		singleRate/1e6, workers, allRate/1e6, speedup)

	// With independent columns and a sharded pool we expect clearly super-linear
	// aggregate throughput vs a single worker. A serialized implementation would
	// stay ~1x. Use a conservative threshold to avoid flakiness on busy CI.
	if speedup < 2.0 {
		t.Errorf("cross-column read did not scale: %.2fx aggregate speedup for %d workers "+
			"(expected >= 2x); a shared lock is likely serializing columns", speedup, workers)
	}
}

func scanAll(t *testing.T, c *Column, repeats int) {
	t.Helper()
	for r := 0; r < repeats; r++ {
		var n int64
		var sum int64
		if err := c.ScanI64(minInt64, maxInt64, func(ts, vals []int64) bool {
			n += int64(len(ts))
			for _, v := range vals {
				sum += v
			}
			return true
		}); err != nil {
			t.Fatal(err)
		}
		_ = sum
	}
}

// TestConcurrentMixedColumns exercises writers and readers on disjoint columns
// at the same time, primarily as a race-detector target (run with -race). It
// also confirms the readers see a consistent count throughout.
func TestConcurrentMixedColumns(t *testing.T) {
	const w = 4
	const writePts = 500_000
	const readFill = 500_000

	db, _ := Open(t.TempDir(), Options{Compress: false})
	defer db.Close()

	readCols := fillCols(t, db, "r", w, readFill)
	writeCols := make([]*Column, w)
	for i := 0; i < w; i++ {
		c, _ := db.Column("b", fmt.Sprintf("w%d", i), Schema{Scheme: SchemeDecimal, Decimals: 2})
		writeCols[i] = c
	}

	base := int64(1_700_000_000_000)
	var wg sync.WaitGroup
	for i := 0; i < w; i++ {
		wg.Add(1)
		go func(c *Column) {
			defer wg.Done()
			for j := int64(0); j < writePts; j++ {
				if err := c.PutI64(base+j*20, 2000+(j%13)); err != nil {
					t.Error(err)
					return
				}
			}
		}(writeCols[i])
	}
	for i := 0; i < w; i++ {
		wg.Add(1)
		go func(c *Column) {
			defer wg.Done()
			for r := 0; r < 10; r++ {
				var n int64
				c.ScanI64(minInt64, maxInt64, func(ts, vals []int64) bool {
					n += int64(len(ts))
					return true
				})
				if n != readFill {
					t.Errorf("concurrent read saw %d points, want %d", n, readFill)
					return
				}
			}
		}(readCols[i])
	}
	wg.Wait()
}

// BenchmarkParallelReadColumns measures aggregate read throughput across
// GOMAXPROCS goroutines each scanning its own column, exercising the sharded
// FilePool under contention.
func BenchmarkParallelReadColumns(b *testing.B) {
	const w = 8
	const fill = 500_000
	db, _ := Open(b.TempDir(), Options{Compress: false})
	defer db.Close()
	cols := fillCols(b, db, "c", w, fill)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for k := 0; k < w; k++ {
			wg.Add(1)
			go func(c *Column) {
				defer wg.Done()
				var s int64
				c.ScanI64(minInt64, maxInt64, func(ts, vals []int64) bool {
					for _, v := range vals {
						s += v
					}
					return true
				})
				_ = s
			}(cols[k])
		}
		wg.Wait()
	}
	b.StopTimer()
	b.ReportMetric(float64(int64(w)*fill*int64(b.N))/b.Elapsed().Seconds(), "points/s")
}
