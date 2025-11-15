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
	"slices"
	"sort"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/xiter"
)

func TestM4(t *testing.T) {
	points := slices.Values([]Point[UnixMilli, I64]{
		{X: 0, Y: 100},
		{X: 1, Y: 230},
		{X: 2, Y: 30},
		{X: 3, Y: 345},
		{X: 4, Y: 2340},
		{X: 5, Y: 65},
		{X: 6, Y: 456},

		{X: 7, Y: 4524},
		{X: 8, Y: 234},
		{X: 9, Y: 567},
		{X: 10, Y: 2432},
		{X: 11, Y: 2},

		{X: 12, Y: 345},
		{X: 13, Y: 3},
		{X: 14, Y: 6},
		{X: 15, Y: 456},
		{X: 16, Y: 67},
		{X: 17, Y: 456},
	})

	expected := []Point[I64, I64]{
		{X: 0, Y: 100},
		{X: 2, Y: 30},
		{X: 4, Y: 2340},
		{X: 5, Y: 65},

		{X: 6, Y: 456},
		{X: 7, Y: 4524},
		{X: 11, Y: 2},

		{X: 12, Y: 345},
		{X: 13, Y: 3},
		{X: 15, Y: 456},
		{X: 17, Y: 456},
	}

	ds := M4[I64](points, NewRange(0, 17, time.UTC), 3)
	res := xiter.Map[Point[UnixMilli, I64], Point[I64, I64]](func(p Point[UnixMilli, I64]) Point[I64, I64] {
		return Point[I64, I64]{X: I64(p.X), Y: p.Y}
	}, ds)
	downsampled := slices.Collect(res)

	if !reflect.DeepEqual(downsampled, expected) {
		t.Fatalf("expected:\n%+v\nbut got:\n%+v", expected, downsampled)
	}
}

func TestM4Order(t *testing.T) {
	rd := rand.New(rand.NewSource(1234))
	for i := 0; i < 1000; i++ {
		pts := make(Series[I64], rd.Intn(10_000))
		for i := range pts {
			pts[i].X = UnixMilli(rand.Intn(100_000) - 50_000)
			pts[i].Y = I64(rand.Intn(100_000) - 50_000)
		}

		sort.Sort(pts)

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

func TestM4Table(t *testing.T) {
	type args struct {
		ts    Series[I64]
		width int
	}
	tests := []struct {
		name string
		args args
		want []Point[I64, I64]
	}{
		{
			name: "less than window size",
			args: args{
				ts: Series[I64]{
					{X: 0, Y: 10},
					{X: 1, Y: 20},
					{X: 2, Y: 30},
				},
				width: 1000,
			},
			want: []Point[I64, I64]{
				{X: 0, Y: 10},
				{X: 1, Y: 20},
				{X: 2, Y: 30},
			},
		},

		{
			name: "unordered values",
			args: args{
				ts: Series[I64]{
					{X: 0, Y: 100},
					{X: 1, Y: 230},
					{X: 2, Y: 30},
					{X: 3, Y: 345},
					{X: 4, Y: 2340},
					{X: 5, Y: 65},

					{X: 6, Y: 456},
					{X: 7, Y: 4524},
					{X: 8, Y: 234},
					{X: 9, Y: 567},
					{X: 10, Y: 2432},
					{X: 11, Y: 2},

					{X: 12, Y: 345},
					{X: 13, Y: 3},
					{X: 14, Y: 6},
					{X: 15, Y: 456},
					{X: 16, Y: 67},
					{X: 17, Y: 456},
				},
				width: 3,
			},
			want: []Point[I64, I64]{
				{X: 0, Y: 100},
				{X: 2, Y: 30},
				{X: 4, Y: 2340},
				{X: 5, Y: 65},

				{X: 6, Y: 456},
				{X: 7, Y: 4524},
				{X: 11, Y: 2},

				{X: 12, Y: 345},
				{X: 13, Y: 3},
				{X: 15, Y: 456},
				{X: 17, Y: 456}, // TODO same value but different indices: is this correct?
			},
		},
		{
			name: "simple even case",
			args: args{
				ts: Series[I64]{
					{X: 0, Y: 10},
					{X: 1, Y: 20},
					{X: 2, Y: 30},
					{X: 3, Y: 40},
					{X: 4, Y: 50},
					{X: 5, Y: 60},

					{X: 6, Y: 70},
					{X: 7, Y: 80},
					{X: 8, Y: 90},
					{X: 9, Y: 100},
					{X: 10, Y: 110},
					{X: 11, Y: 120},

					{X: 12, Y: 130},
					{X: 13, Y: 140},
					{X: 14, Y: 150},
					{X: 15, Y: 160},
					{X: 16, Y: 170},
					{X: 17, Y: 180},
				},
				width: 3,
			},
			want: []Point[I64, I64]{
				{X: 0, Y: 10},
				{X: 5, Y: 60},

				{X: 6, Y: 70},
				{X: 11, Y: 120},

				{X: 12, Y: 130},
				{X: 17, Y: 180},
			},
		},

		{
			name: "series with hole",
			args: args{
				ts: Series[I64]{
					{X: 0, Y: 10},
					{X: 1, Y: 20},
					{X: 2, Y: 30},
					{X: 3, Y: 40},
					{X: 4, Y: 50},
					{X: 5, Y: 60},

					// 6-11 is missing

					{X: 12, Y: 130},
					{X: 13, Y: 140},
					{X: 14, Y: 150},
					{X: 15, Y: 160},
					{X: 16, Y: 170},
					{X: 17, Y: 180},
				},
				width: 3,
			},
			want: []Point[I64, I64]{
				{X: 0, Y: 10},
				{X: 5, Y: 60},

				// the middle bucket must be missing

				{X: 12, Y: 130},
				{X: 17, Y: 180},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from := tt.args.ts[0].X
			to := tt.args.ts[len(tt.args.ts)-1].X
			ds := M4[I64](slices.Values(tt.args.ts), NewRange(from, to, time.UTC), 3)
			res := xiter.Map[Point[UnixMilli, I64], Point[I64, I64]](func(p Point[UnixMilli, I64]) Point[I64, I64] {
				return Point[I64, I64]{X: I64(p.X), Y: p.Y}
			}, ds)
			got := slices.Collect(res)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("M4() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
