// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// TestScale measures end-to-end insert and read throughput, on-disk size, and
// peak heap for a monotonic time series, at one or more scales. It is a
// measurement harness rather than a pass/fail test: it always passes but logs a
// results table.
//
// Scales are controlled by the TSDB_SCALES env var (comma-separated counts).
// Default is "1000000". To include the billion-point run:
//
//	TSDB_SCALES=1000000,1000000000 go test -run TestScale -v -timeout 60m ./pkg/ndb/tsdb/
//
// The harness asserts the invariant that matters for burst ingest: peak heap
// growth during insert stays bounded (a small constant), independent of the
// number of points, proving the monotonic append fast path does not accumulate
// data in memory.
func TestScale(t *testing.T) {
	scales := []int64{1_000_000}
	if s := os.Getenv("TSDB_SCALES"); s != "" {
		scales = scales[:0]
		for _, part := range splitComma(s) {
			n, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				t.Fatalf("bad TSDB_SCALES %q: %v", part, err)
			}
			scales = append(scales, n)
		}
	}

	type row struct {
		n            int64
		insDur       time.Duration
		insRate      float64
		rdDur        time.Duration
		rdRate       float64
		fileBytes    int64
		bytesPerPt   float64
		heapGrowthMB float64
	}
	var rows []row

	for _, n := range scales {
		dir, err := os.MkdirTemp("", "tsdb-scale-")
		if err != nil {
			t.Fatal(err)
		}

		db, err := Open(dir, Options{Compress: false})
		if err != nil {
			t.Fatal(err)
		}
		c, err := db.Column("scale", "sig", Schema{Scheme: SchemeDecimal, Decimals: 2})
		if err != nil {
			t.Fatal(err)
		}

		var m0 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m0)

		base := int64(1_700_000_000_000)
		var peakHeap uint64
		// TSDB_FLAT=1 writes a constant value (exercises the constant-value block
		// optimization); default writes an oscillating value.
		flat := os.Getenv("TSDB_FLAT") == "1"
		start := time.Now()
		for i := int64(0); i < n; i++ {
			// 50 Hz equidistant timestamps.
			v := int64(2000 + (i % 13)) // oscillating, small delta
			if flat {
				v = 2000 // constant → constant-value block encoding
			}
			if err := c.PutI64(base+i*20, v); err != nil {
				t.Fatal(err)
			}
			if i&((1<<20)-1) == 0 { // sample heap every ~1M points
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				if m.HeapInuse > peakHeap {
					peakHeap = m.HeapInuse
				}
			}
		}
		if err := c.Flush(); err != nil {
			t.Fatal(err)
		}
		insDur := time.Since(start)

		// read: full scan summing values (columnar batch path)
		var sink int64
		var got int64
		rstart := time.Now()
		if err := c.ScanI64(minInt64, maxInt64, func(ts []int64, vals []int64) bool {
			got += int64(len(ts))
			for _, v := range vals {
				sink += v
			}
			return true
		}); err != nil {
			t.Fatal(err)
		}
		rdDur := time.Since(rstart)
		_ = sink
		if got != n {
			t.Fatalf("scale %d: read back %d points, want %d", n, got, n)
		}

		fileBytes := dirSize(t, dir)

		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		heapGrowth := float64(int64(peakHeap)-int64(m0.HeapInuse)) / (1 << 20)

		rows = append(rows, row{
			n:            n,
			insDur:       insDur,
			insRate:      float64(n) / insDur.Seconds(),
			rdDur:        rdDur,
			rdRate:       float64(n) / rdDur.Seconds(),
			fileBytes:    fileBytes,
			bytesPerPt:   float64(fileBytes) / float64(n),
			heapGrowthMB: heapGrowth,
		})

		db.Close()
		os.RemoveAll(dir)

		// invariant: peak heap growth during insert must be bounded (well under
		// 512 MiB) regardless of n — proving constant-memory ingest.
		if heapGrowth > 512 {
			t.Errorf("scale %d: peak heap grew %.0f MiB during insert; ingest is not constant-memory", n, heapGrowth)
		}
	}

	// print results table
	t.Log("")
	t.Logf("%-14s | %-12s | %-14s | %-12s | %-14s | %-12s | %-10s | %-12s",
		"points", "insert", "insert pts/s", "read", "read pts/s", "on-disk", "bytes/pt", "peak heap MiB")
	t.Logf("%s", "---------------+--------------+----------------+--------------+----------------+--------------+------------+-------------")
	for _, r := range rows {
		t.Logf("%-14s | %-12s | %-14s | %-12s | %-14s | %-12s | %-10.2f | %-12.1f",
			humanCount(r.n),
			roundDur(r.insDur),
			humanRate(r.insRate),
			roundDur(r.rdDur),
			humanRate(r.rdRate),
			humanBytes(r.fileBytes),
			r.bytesPerPt,
			r.heapGrowthMB,
		)
	}
}

func splitComma(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ',' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func dirSize(t *testing.T, dir string) int64 {
	t.Helper()
	var total int64
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return total
}

func humanCount(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.0fB", float64(n)/1e9)
	case n >= 1_000_000:
		return fmt.Sprintf("%.0fM", float64(n)/1e6)
	case n >= 1_000:
		return fmt.Sprintf("%.0fk", float64(n)/1e3)
	default:
		return strconv.FormatInt(n, 10)
	}
}

func humanRate(r float64) string {
	switch {
	case r >= 1e6:
		return fmt.Sprintf("%.1fM/s", r/1e6)
	case r >= 1e3:
		return fmt.Sprintf("%.0fk/s", r/1e3)
	default:
		return fmt.Sprintf("%.0f/s", r)
	}
}

func humanBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.2f GiB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KiB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func roundDur(d time.Duration) string {
	if d >= time.Second {
		return d.Round(10 * time.Millisecond).String()
	}
	return d.Round(time.Microsecond).String()
}
