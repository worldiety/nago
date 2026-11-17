// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"

	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// BenchmarkN22ScriptSumsBlocks is the measurement of L-N22-25 (wokoda N22, threshold ≥ 55 M points/s, M22c V5 44.5): a
// script of the vm sums the blocks of a series, which a host function hands it with one callback per block
// (n22h.FoldFloats, SumFold). Sources: a Go slice of blocks of 4 096 and tsdb ScanF64 of a column like M22 power_kw
// (1 decimal, 1-s grid, compression), N22_BENCH_POINTS points (default 1 000 000). Modes: the fast writer with
// Writer.Float64s (fast), Callback.Call with []float64 (typed) and with a []any per block (as before N22). It runs only
// with -bench, e.g. against nago f4aca80c5 with the overlay of testdata/n22, which writes the list element by element:
//
//	N22_BENCH_POINTS=10000000 go test -run '^$' -bench BenchmarkN22ScriptSumsBlocks -count 6 ./pkg/gostrict/
//	cd pkg/gostrict && go test -overlay testdata/n22/overlay.json -run '^$' -bench 'BenchmarkN22ScriptSumsBlocks/.*/(fast|any)' .
func BenchmarkN22ScriptSumsBlocks(b *testing.B) {
	n := 1_000_000
	if s := os.Getenv("N22_BENCH_POINTS"); s != "" {
		var err error
		if n, err = strconv.Atoi(s); err != nil || n <= 0 {
			b.Fatalf("N22_BENCH_POINTS=%q is no positive number", s)
		}
	}
	vals := n22PowerSeries(n)
	var blocks [][]float64
	for lo := 0; lo < n; lo += 4096 {
		blocks = append(blocks, vals[lo:min(lo+4096, n)])
	}
	var want float64
	for _, v := range vals {
		want += v
	}

	dir := b.TempDir()
	db, err := tsdb.Open(dir, tsdb.Options{FilePool: ndb.NewFilePool(100_000), Compress: true})
	if err != nil {
		b.Fatal(err)
	}
	c, err := db.Column("n22", "power_kw", tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1})
	if err != nil {
		b.Fatal(err)
	}
	for i, v := range vals {
		if err := c.PutF64(1_735_689_600_000+int64(i)*1000, v); err != nil {
			b.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		b.Fatal(err)
	}
	if db, err = tsdb.Open(dir, tsdb.Options{FilePool: ndb.NewFilePool(100_000), Compress: true}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = db.Close() })
	if c, _, err = db.LookupColumn("n22", "power_kw"); err != nil || c == nil {
		b.Fatalf("lookup: %v", err)
	}
	scan := func(fn func([]float64) bool) error {
		return c.ScanF64(math.MinInt64, math.MaxInt64, func(_ []int64, v []float64) bool { return fn(v) })
	}

	bundle := n22Bundle(b)
	for _, src := range []string{"slice", "tsdb"} {
		for _, mode := range []n22Mode{n22Fast, n22Typed, n22Any} {
			b.Run(src+"/"+map[n22Mode]string{n22Fast: "fast", n22Typed: "typed", n22Any: "any"}[mode], func(b *testing.B) {
				h := &n22Host{mode: mode, blocks: blocks}
				if src == "tsdb" {
					h.scan = scan
				}
				in, err := new(vm.Engine).Load(context.Background(), bundle, h.funcs())
				if err != nil {
					b.Fatal(err)
				}
				defer in.Close()
				ctx := context.Background()
				var got any
				for b.Loop() {
					res, err := in.Call(ctx, "SumFold")
					if err != nil || h.callErr != nil {
						b.Fatalf("%v %v", err, h.callErr)
					}
					got = res[0]
				}
				// the values have one decimal, so the sum of the script and the one in Go may differ in the last bits
				if g, _ := got.(float64); math.Abs(g-want) > 1e-6*math.Abs(want) {
					b.Fatalf("the script sums %v, Go %v", got, want)
				}
				b.ReportMetric(float64(n*b.N)/b.Elapsed().Seconds(), "points/s")
			})
		}
	}
}

// n22PowerSeries is power of M22d (pkg/ndb/tsdb n22Values "power": a random walk with a daily pattern), with 1 decimal.
func n22PowerSeries(n int) []float64 {
	var day [144]float64
	for i := range day {
		day[i] = math.Sin(2*math.Pi*float64(i)/144 - math.Pi/2)
	}
	r := rand.New(rand.NewPCG(2, 0x9e3779b97f4a7c15))
	x := 1500.0
	v := make([]float64, n)
	for i := range v {
		x += 0.01*(1500-x) + 40*r.NormFloat64()
		y := min(max(x+300*day[i%144], 0), 3600)
		v[i] = math.Round(y*10) / 10
	}
	return v
}
