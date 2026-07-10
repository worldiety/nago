// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import "testing"

func benchColumn(b *testing.B, n int64) *Column {
	b.Helper()
	db, err := Open(b.TempDir(), Options{Compress: false})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { db.Close() })
	c, _ := db.Column("bench", "sig", Schema{Scheme: SchemeDecimal, Decimals: 2})
	base := int64(1_700_000_000_000)
	for i := int64(0); i < n; i++ {
		// oscillating value with small relative change (delta ~ few units)
		c.PutI64(base+i*20, 2000+(i%13))
	}
	if err := c.Flush(); err != nil {
		b.Fatal(err)
	}
	return c
}

// BenchmarkScanI64 measures the columnar batch read path (the billion-element
// target). Reports points/sec via b.N iterations over a fixed dataset.
func BenchmarkScanI64(b *testing.B) {
	const n = 1_000_000
	c := benchColumn(b, n)
	b.ResetTimer()
	var sink int64
	for i := 0; i < b.N; i++ {
		_ = c.ScanI64(minInt64, maxInt64, func(ts []int64, vals []int64) bool {
			for _, v := range vals {
				sink += v
			}
			return true
		})
	}
	b.StopTimer()
	b.ReportMetric(float64(n*int64(b.N))/b.Elapsed().Seconds(), "points/s")
	_ = sink
}

// BenchmarkPutI64 measures the write path into the head log.
func BenchmarkPutI64(b *testing.B) {
	db, _ := Open(b.TempDir(), Options{})
	defer db.Close()
	c, _ := db.Column("bench", "sig", Schema{Scheme: SchemeDecimal, Decimals: 2})
	base := int64(1_700_000_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.PutI64(base+int64(i)*20, 2000+int64(i%13))
	}
}

func BenchmarkIterI64(b *testing.B) {
	const n = 1_000_000
	c := benchColumn(b, n)
	b.ResetTimer()
	var sink int64
	for i := 0; i < b.N; i++ {
		for p := range c.IterI64(minInt64, maxInt64) {
			sink += int64(p.Y)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(n*int64(b.N))/b.Elapsed().Seconds(), "points/s")
	_ = sink
}

func BenchmarkIterF64(b *testing.B) {
	const n = 1_000_000
	c := benchColumn(b, n)
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		for p := range c.IterF64(minInt64, maxInt64) {
			sink += float64(p.Y)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(n*int64(b.N))/b.Elapsed().Seconds(), "points/s")
	_ = sink
}
