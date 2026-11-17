// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"os"
	"strconv"
	"testing"

	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The benchmarks of the measurement of L-N22-25 (wokoda N22), on the data of M22 and M22d. They run only with -bench,
// e.g. in full size, alternating with nago f4aca80c5 on the same machine:
//
//	N22_BENCH_POINTS=50000000 go test -run '^$' -bench 'BenchmarkN22LongSeries' -count 6 -timeout 60m ./pkg/ndb/tsdb/
//	go test -run '^$' -bench 'BenchmarkN22DecodeBlocks' -count 6 ./pkg/ndb/tsdb/
//
// Each benchmark writes its column once per run of the process, closes the database, opens it again and reads the
// column once before it measures, so that it measures warm reads from the page cache. A sum goes into a variable local
// to the callback and is added to the total once per block (L-N22-21: the instrument of M22c summed into a captured
// variable).

// n22BenchInt reads a positive integer from the environment, def if it is not set.
func n22BenchInt(b *testing.B, name string, def int) int {
	s := os.Getenv(name)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		b.Fatalf("%s=%q is no positive number", name, s)
	}
	return n
}

// n22BenchColumn writes vals of kind in the 1-s grid into a new column with compression, closes the database and
// returns the column of a new open, read once.
func n22BenchColumn(b *testing.B, name string, decimals uint8, vals []int64) *tsdb.Column {
	b.Helper()
	dir := b.TempDir()
	opts := tsdb.Options{FilePool: ndb.NewFilePool(100_000), Compress: true}
	db, err := tsdb.Open(dir, opts)
	if err != nil {
		b.Fatal(err)
	}
	c, err := db.Column("n22", name, tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: decimals})
	if err != nil {
		b.Fatal(err)
	}
	for i, v := range vals {
		if err := c.PutI64(n22Base+int64(i)*1000, v); err != nil {
			b.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		b.Fatal(err)
	}
	opts.FilePool = ndb.NewFilePool(100_000)
	db, err = tsdb.Open(dir, opts)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = db.Close() })
	c, ok, err := db.LookupColumn("n22", name)
	if err != nil || !ok {
		b.Fatalf("lookup: %v %v", ok, err)
	}
	n := 0
	if err := c.ScanI64(minTS, maxTS, func(ts, _ []int64) bool { n += len(ts); return true }); err != nil || n != len(vals) {
		b.Fatalf("the warm-up read sees %d of %d points: %v", n, len(vals), err)
	}
	return c
}

// BenchmarkN22LongSeries reads a long series: power with 1 decimal in the 1-s grid with compression (L-N22-25,
// thresholds ScanI64 counting ≥ 600 M points/s, ScanF64 ≥ 180 M points/s), N22_BENCH_POINTS points (default 1 000 000,
// L-N22-25 50 000 000). ScanI64/count only counts, ScanI64/sum and ScanF64/sum sum each block.
func BenchmarkN22LongSeries(b *testing.B) {
	n := n22BenchInt(b, "N22_BENCH_POINTS", 1_000_000)
	c := n22BenchColumn(b, "power", 1, n22Values("power", n))
	report := func(b *testing.B, seen int) {
		if seen != n*b.N {
			b.Fatalf("%d points read, want %d", seen, n*b.N)
		}
		b.ReportMetric(float64(seen)/b.Elapsed().Seconds(), "points/s")
	}
	b.Run("ScanI64/count", func(b *testing.B) {
		seen := 0
		for b.Loop() {
			if err := c.ScanI64(minTS, maxTS, func(ts, vals []int64) bool {
				seen += len(vals)
				return true
			}); err != nil {
				b.Fatal(err)
			}
		}
		report(b, seen)
	})
	b.Run("ScanI64/sum", func(b *testing.B) {
		seen, total := 0, int64(0)
		for b.Loop() {
			if err := c.ScanI64(minTS, maxTS, func(ts, vals []int64) bool {
				var s int64
				for _, v := range vals {
					s += v
				}
				total += s
				seen += len(vals)
				return true
			}); err != nil {
				b.Fatal(err)
			}
		}
		report(b, seen)
		_ = total
	})
	b.Run("ScanF64/sum", func(b *testing.B) {
		seen, total := 0, 0.0
		for b.Loop() {
			if err := c.ScanF64(minTS, maxTS, func(ts []int64, vals []float64) bool {
				var s float64
				for _, v := range vals {
					s += v
				}
				total += s
				seen += len(vals)
				return true
			}); err != nil {
				b.Fatal(err)
			}
		}
		report(b, seen)
		_ = total
	})
}

// BenchmarkN22DecodeBlocks reads N22_BENCH_BLOCKS blocks of 4 096 points (default 64) of each kind of data of M22d
// with ScanI64 and a sum per block; the status goes as its ids into a decimal column, as M22d measured it. It measures
// the read of a block through the surface of tsdb (frame, CRC, mask and the values), not the codec alone, whose names
// are inside tsdb.
func BenchmarkN22DecodeBlocks(b *testing.B) {
	blocks := n22BenchInt(b, "N22_BENCH_BLOCKS", 64)
	n := blocks * 4096
	for _, k := range n22Kinds {
		b.Run(k.name, func(b *testing.B) {
			c := n22BenchColumn(b, k.name, k.decimals, n22Values(k.name, n))
			seen, total := 0, int64(0)
			for b.Loop() {
				if err := c.ScanI64(minTS, maxTS, func(ts, vals []int64) bool {
					var s int64
					for _, v := range vals {
						s += v
					}
					total += s
					seen += len(vals)
					return true
				}); err != nil {
					b.Fatal(err)
				}
			}
			if seen != n*b.N {
				b.Fatalf("%d points read, want %d", seen, n*b.N)
			}
			b.ReportMetric(float64(seen)/b.Elapsed().Seconds(), "points/s")
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(blocks*b.N), "ns/block")
			_ = total
		})
	}
}
