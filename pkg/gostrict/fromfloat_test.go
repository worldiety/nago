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
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// TestFromFloat compares text.FromFloat in goja with Go, especially for ties, which toFixed rounds differently.
func TestFromFloat(t *testing.T) {
	cache := gostrict.NewCache(1)
	b, diags, err := cache.Compile("llm.local/s", script(`package s

import "go.wdy.de/nago/pkg/gostrict/std/text"

func F(f float64, d int32) string { return text.FromFloat(f, d) }
`))
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	in := mustLoad(t, goEngine, b)

	type c struct {
		f float64
		d int32
	}
	cases := []c{
		{0.125, 2}, {0.375, 2}, {2.5, 0}, {3.5, 0}, {-2.5, 0}, {-0.125, 2}, {0.135, 2}, {1.005, 2},
		{0, 2}, {math.Copysign(0, -1), 2}, {-0.001, 2}, {1e21, 2}, {-1.5e300, 1}, {123456789.125, 2},
		{math.NaN(), 2}, {math.Inf(1), 2}, {math.Inf(-1), 0}, {5e-324, 3}, {0.5, 0}, {1.5, 0},
		{3.14159, -3}, {1.0 / 3, 200}, {9.995, 2}, {0.0625, 3}, {2.675, 2},
	}
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 2000 {
		// many ties: multiples of powers of 1/2
		f := float64(rnd.IntN(100000)-50000) / float64(int(1)<<rnd.IntN(12))
		cases = append(cases, c{f, int32(rnd.IntN(8))})
		cases = append(cases, c{rnd.NormFloat64() * math.Pow(10, float64(rnd.IntN(30)-10)), int32(rnd.IntN(12))})
	}

	for _, tc := range cases {
		want := text.FromFloat(tc.f, tc.d)
		res, err := in.Call(context.Background(), "F", tc.f, tc.d)
		if err != nil {
			t.Fatal(err)
		}
		if got := res[0].(string); got != want {
			t.Fatalf("FromFloat(%v, %d): goja %q, go %q", tc.f, tc.d, got, want)
		}
	}
}
