// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"math/rand"
	"reflect"
	"testing"

	"go.wdy.de/nago/pkg/timeseries"
)

// M4's own algorithmic correctness is proven
// independently in pkg/timeseries/m4_test.go against a fixed-grid reference and
// under a constant-memory streaming test. The tests here prove the complementary
// tsdb-side property: the tsdb column iterator is a faithful M4 data source, so
// M4(tsdb.IterI64) is identical to M4 over the same data delivered by a plain
// in-memory Series — regardless of tsdb's internal storage state (append buffer,
// sealed chunks, or an out-of-order head correction). If the iterator dropped,
// duplicated, reordered or mis-valued any point, the two M4 results would differ.

func seriesSeq(ts, vals []int64) func(func(timeseries.Point[timeseries.UnixMilli, timeseries.I64]) bool) {
	return func(yield func(timeseries.Point[timeseries.UnixMilli, timeseries.I64]) bool) {
		for i := range ts {
			if !yield(timeseries.Point[timeseries.UnixMilli, timeseries.I64]{
				X: timeseries.UnixMilli(ts[i]), Y: timeseries.I64(vals[i]),
			}) {
				return
			}
		}
	}
}

func collectM4(seq func(func(timeseries.Point[timeseries.UnixMilli, timeseries.I64]) bool)) []timeseries.Point[timeseries.UnixMilli, timeseries.I64] {
	var out []timeseries.Point[timeseries.UnixMilli, timeseries.I64]
	for p := range seq {
		out = append(out, p)
	}
	return out
}

func TestM4OverTSDBIteratorMatchesInMemory(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	// oscillating, holed series so per-bucket first/last/min/max genuinely differ
	var ts, vals []int64
	base := int64(1_700_000_000_000)
	tprev := base
	for i := 0; i < 5000; i++ {
		step := int64(20)
		if rng.Intn(50) == 0 {
			step += int64(rng.Intn(500))
		}
		tprev += step
		ts = append(ts, tprev)
		vals = append(vals, int64(rng.Intn(1000))-500)
	}
	tMin, tMax := ts[0], ts[len(ts)-1]

	for _, tc := range []struct {
		name  string
		flush bool
	}{
		{"buffered (unflushed)", false},
		{"sealed (flushed)", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := openTestDB(t)
			c, err := db.Column("plant", "signal", Schema{Scheme: SchemeDecimal, Decimals: 0})
			if err != nil {
				t.Fatal(err)
			}
			for i := range ts {
				if err := c.PutI64(ts[i], vals[i]); err != nil {
					t.Fatal(err)
				}
			}
			if tc.flush {
				if err := c.Flush(); err != nil {
					t.Fatal(err)
				}
			}

			for _, width := range []int{1, 2, 10, 100, 1000, 6000} {
				interval := timeseries.NewRange(timeseries.UnixMilli(tMin), timeseries.UnixMilli(tMax), timeUTC())
				fromTSDB := collectM4(timeseries.M4(c.IterI64(tMin, tMax+1), interval, width))
				fromMem := collectM4(timeseries.M4(seriesSeq(ts, vals), interval, width))
				if !reflect.DeepEqual(fromTSDB, fromMem) {
					t.Fatalf("width=%d: M4(tsdb) != M4(in-memory)\n tsdb=%v\n mem=%v", width, fromTSDB, fromMem)
				}
			}
		})
	}
}

// TestM4OverTSDBAfterOutOfOrderCorrection proves the iterator stays a faithful
// M4 source when a correction is pending in the head (not yet compacted): M4 over
// the tsdb column must equal M4 over the logically corrected in-memory series.
func TestM4OverTSDBAfterOutOfOrderCorrection(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})

	base := int64(1_700_000_000_000)
	n := 2000
	ts := make([]int64, n)
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		ts[i] = base + int64(i)*20
		vals[i] = int64(i % 11)
		c.PutI64(ts[i], vals[i])
	}
	c.Flush() // seal

	// out-of-order corrections routed through the head
	for idx, v := range map[int]int64{500: 99999, 1234: -99999, 1900: 42} {
		vals[idx] = v
		if err := c.PutI64(ts[idx], v); err != nil {
			t.Fatal(err)
		}
	}

	tMin, tMax := ts[0], ts[n-1]
	for _, width := range []int{10, 100, 500} {
		interval := timeseries.NewRange(timeseries.UnixMilli(tMin), timeseries.UnixMilli(tMax), timeUTC())
		fromTSDB := collectM4(timeseries.M4(c.IterI64(tMin, tMax+1), interval, width))
		fromMem := collectM4(timeseries.M4(seriesSeq(ts, vals), interval, width))
		if !reflect.DeepEqual(fromTSDB, fromMem) {
			t.Fatalf("width=%d: corrected M4(tsdb) != M4(in-memory)\n tsdb=%v\n mem=%v", width, fromTSDB, fromMem)
		}
	}
}
