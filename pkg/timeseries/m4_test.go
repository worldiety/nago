// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeseries

import (
	"math"
	"math/rand"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"testing"
	"time"
)

// refM4 is an independent, deliberately simple reference implementation of the
// M4 algorithm
//
//   - the time span [tMin,tMax] is divided into `width` FIXED equal-width grid
//     buckets of width Δ = (tMax-tMin)/width;
//   - a point at time t belongs to bucket floor((t-tMin)/Δ), clamped into
//     [0,width-1] (so t == tMax lands in the last bucket);
//   - each non-empty bucket contributes the points with min-time (first),
//     max-time (last), min-value and max-value, de-duplicated, in ascending
//     time order.
//
// This is written for clarity, not speed, and does NOT share code with the
// streaming implementation under test, so it cannot re-encode the same bug.
func refM4(pts []Point[UnixMilli, I64], tMin, tMax UnixMilli, width int) []Point[UnixMilli, I64] {
	if width <= 0 {
		return append([]Point[UnixMilli, I64](nil), pts...)
	}
	delta := (tMax - tMin) / UnixMilli(width)
	if delta <= 0 {
		return append([]Point[UnixMilli, I64](nil), pts...)
	}

	type bucket struct {
		first, last, min, max Point[UnixMilli, I64]
		has                   bool
	}
	buckets := make([]bucket, width)

	for _, p := range pts {
		bi := int64((p.X - tMin) / delta)
		if bi < 0 {
			bi = 0
		}
		if bi >= int64(width) {
			bi = int64(width) - 1
		}
		b := &buckets[bi]
		if !b.has {
			b.first, b.last, b.min, b.max = p, p, p, p
			b.has = true
			continue
		}
		b.last = p
		if p.Y < b.min.Y {
			b.min = p
		}
		if p.Y > b.max.Y {
			b.max = p
		}
	}

	var out []Point[UnixMilli, I64]
	for i := range buckets {
		b := &buckets[i]
		if !b.has {
			continue
		}
		var cand [4]Point[UnixMilli, I64]
		cand[0] = b.first
		if b.min.X <= b.max.X {
			cand[1], cand[2] = b.min, b.max
		} else {
			cand[1], cand[2] = b.max, b.min
		}
		cand[3] = b.last
		var prev UnixMilli
		for j, c := range cand {
			if j > 0 && c.X == prev {
				continue
			}
			prev = c.X
			out = append(out, c)
		}
	}
	return out
}

func runM4(pts []Point[UnixMilli, I64], tMin, tMax UnixMilli, width int) []Point[UnixMilli, I64] {
	return slices.Collect(M4[I64](slices.Values(pts), NewRange(tMin, tMax, time.UTC), width))
}

// TestM4MatchesPaperReference is the primary correctness proof: for a wide range
// of inputs (regular, irregular, holes, spikes, random) the streaming M4 must
// produce exactly the same result as the independent fixed-grid paper reference.
func TestM4MatchesPaperReference(t *testing.T) {
	mk := func(xs, ys []int64) []Point[UnixMilli, I64] {
		out := make([]Point[UnixMilli, I64], len(xs))
		for i := range xs {
			out[i] = Point[UnixMilli, I64]{X: UnixMilli(xs[i]), Y: I64(ys[i])}
		}
		return out
	}

	cases := []struct {
		name  string
		pts   []Point[UnixMilli, I64]
		width int
	}{
		{
			"simple even, w=3",
			mk([]int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
				[]int64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180}),
			3,
		},
		{
			"oscillating, w=3",
			mk([]int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
				[]int64{100, 230, 30, 345, 2340, 65, 456, 4524, 234, 567, 2432, 2, 345, 3, 6, 456, 67, 456}),
			3,
		},
		{
			"hole in the middle, w=3",
			mk([]int64{0, 1, 2, 3, 4, 5, 12, 13, 14, 15, 16, 17},
				[]int64{10, 20, 30, 40, 50, 60, 130, 140, 150, 160, 170, 180}),
			3,
		},
		{
			"irregular + spikes, w=2",
			mk([]int64{0, 1, 2, 5, 10, 11, 19, 20},
				[]int64{0, 9, 3, 1, 8, 2, 7, 4}),
			2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tMin := tc.pts[0].X
			tMax := tc.pts[len(tc.pts)-1].X
			got := runM4(tc.pts, tMin, tMax, tc.width)
			want := refM4(tc.pts, tMin, tMax, tc.width)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("M4 != paper reference\n got:  %+v\n want: %+v", got, want)
			}
		})
	}
}

// TestM4Fuzz cross-checks the streaming M4 against the paper reference on many
// random, sorted inputs and widths.
func TestM4Fuzz(t *testing.T) {
	rd := rand.New(rand.NewSource(2024))
	for iter := 0; iter < 3000; iter++ {
		n := rd.Intn(300)
		pts := make([]Point[UnixMilli, I64], n)
		x := UnixMilli(rd.Intn(1000))
		for i := range pts {
			x += UnixMilli(rd.Intn(6)) // 0 allows duplicate timestamps
			pts[i] = Point[UnixMilli, I64]{X: x, Y: I64(rd.Intn(2000) - 1000)}
		}
		if n == 0 {
			continue
		}
		width := rd.Intn(20) + 1
		tMin, tMax := pts[0].X, pts[len(pts)-1].X
		got := runM4(pts, tMin, tMax, width)
		want := refM4(pts, tMin, tMax, width)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("iter %d n=%d width=%d: mismatch\n pts=%+v\n got=%+v\n want=%+v",
				iter, n, width, pts, got, want)
		}
	}
}

// TestM4PreservesExtremes proves the defining visual property: every bucket's
// minimum and maximum value appear in the output.
func TestM4PreservesExtremes(t *testing.T) {
	rd := rand.New(rand.NewSource(7))
	for iter := 0; iter < 500; iter++ {
		n := rd.Intn(500) + 1
		pts := make([]Point[UnixMilli, I64], n)
		x := UnixMilli(0)
		for i := range pts {
			x += UnixMilli(rd.Intn(5) + 1)
			pts[i] = Point[UnixMilli, I64]{X: x, Y: I64(rd.Intn(10000))}
		}
		width := rd.Intn(30) + 1
		tMin, tMax := pts[0].X, pts[len(pts)-1].X
		delta := (tMax - tMin) / UnixMilli(width)
		if delta <= 0 {
			continue
		}

		// bucket -> set of emitted values
		emitted := map[int64]map[int64]bool{}
		for _, p := range runM4(pts, tMin, tMax, width) {
			bi := int64((p.X - tMin) / delta)
			if bi >= int64(width) {
				bi = int64(width) - 1
			}
			if emitted[bi] == nil {
				emitted[bi] = map[int64]bool{}
			}
			emitted[bi][int64(p.Y)] = true
		}

		// reference per-bucket min/max
		type mm struct{ mn, mx int64 }
		ref := map[int64]*mm{}
		for _, p := range pts {
			bi := int64((p.X - tMin) / delta)
			if bi >= int64(width) {
				bi = int64(width) - 1
			}
			m := ref[bi]
			if m == nil {
				ref[bi] = &mm{int64(p.Y), int64(p.Y)}
				continue
			}
			if int64(p.Y) < m.mn {
				m.mn = int64(p.Y)
			}
			if int64(p.Y) > m.mx {
				m.mx = int64(p.Y)
			}
		}
		for bi, m := range ref {
			if !emitted[bi][m.mn] {
				t.Fatalf("iter %d: bucket %d min %d not emitted", iter, bi, m.mn)
			}
			if !emitted[bi][m.mx] {
				t.Fatalf("iter %d: bucket %d max %d not emitted", iter, bi, m.mx)
			}
		}
	}
}

// TestM4Order asserts the output is strictly ascending in time (a property M4
// must always guarantee for sorted input).
func TestM4Order(t *testing.T) {
	rd := rand.New(rand.NewSource(1234))
	for i := 0; i < 1000; i++ {
		pts := make(Series[I64], rd.Intn(10_000))
		for i := range pts {
			pts[i].X = UnixMilli(rand.Intn(100_000) - 50_000)
			pts[i].Y = I64(rand.Intn(100_000) - 50_000)
		}
		sort.Sort(pts)
		if len(pts) == 0 {
			continue
		}

		downsampled := slices.Collect(M4[I64](slices.Values(pts), NewRange(pts[0].X, pts[len(pts)-1].X, time.UTC), rand.Intn(1_000)+1))

		min := UnixMilli(math.MinInt64)
		for i, point := range downsampled {
			if point.X >= min {
				min = point.X
			} else {
				t.Fatalf("expected strict monotonic x values (time), last=%d but current is %d @ index %d", min, point.X, i)
			}
		}
	}
}

// TestM4ConstantMemoryStreaming proves the operational property that matters for
// billions of points: M4 consumes its input lazily (one point at a time, pull
// based) and never materializes the whole series. We feed a virtually unbounded
// generator and confirm M4 pulls exactly as far as needed and holds only O(1)
// state — if it buffered the input this test would run out of memory / never
// return.
func TestM4ConstantMemoryStreaming(t *testing.T) {
	const total = 50_000_000 // 50M points; would be ~1.6 GiB if materialized
	const width = 1000

	tMin := UnixMilli(0)
	tMax := UnixMilli(total - 1)

	var pulled int64
	gen := func(yield func(Point[UnixMilli, I64]) bool) {
		for i := int64(0); i < total; i++ {
			pulled++
			// deterministic oscillation so buckets have real min/max
			v := I64((i * 2654435761) % 1000)
			if !yield(Point[UnixMilli, I64]{X: UnixMilli(i), Y: v}) {
				return
			}
		}
	}

	var m0 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m0)

	var emitted int64
	var lastX UnixMilli = math.MinInt64
	for p := range M4[I64](gen, NewRange(tMin, tMax, time.UTC), width) {
		emitted++
		if p.X <= lastX && emitted > 1 {
			t.Fatalf("non-ascending output at %d", p.X)
		}
		lastX = p.X
	}

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	growth := int64(m1.HeapInuse) - int64(m0.HeapInuse)

	if pulled != total {
		t.Fatalf("generator pulled %d of %d points", pulled, total)
	}
	// at most 4 points per bucket, width buckets
	if emitted > int64(4*width) {
		t.Fatalf("emitted %d points, must be <= %d (4 per bucket)", emitted, 4*width)
	}
	// heap growth must be a small constant, independent of the 50M input size.
	if growth > 32<<20 {
		t.Fatalf("M4 heap grew %d bytes over a 50M-point stream; not constant memory", growth)
	}
	t.Logf("streamed %d points -> %d output points, heap growth %d KiB", total, emitted, growth/1024)
}
