// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"fmt"
	"math"
	"testing"
)

// math.Min and math.Max behave like in Go also with NaN, the infinities and the zeros (L-N1-5). The builtins min and
// max of Go give NaN for any NaN, also with an infinity, and stay so.

const minMaxSrc = `package s

import "math"

// each function returns the result and its sign bit, as the program sees it

func Max(a, b float64) (float64, bool) {
	m := math.Max(a, b)
	return m, math.Signbit(m)
}

func Min(a, b float64) (float64, bool) {
	m := math.Min(a, b)
	return m, math.Signbit(m)
}

func BuiltinMax(a, b float64) (float64, bool) {
	m := max(a, b)
	return m, math.Signbit(m)
}

func BuiltinMin(a, b float64) (float64, bool) {
	m := min(a, b)
	return m, math.Signbit(m)
}

// Zeros makes the zeros and NaN in the program, so that they do not depend on how an engine passes them.
func Zeros() (bool, bool, bool, bool, bool) {
	z := 0.0
	n := -z
	nan := math.NaN()
	return math.Signbit(math.Max(n, z)), math.Signbit(math.Min(z, n)), math.IsInf(math.Max(nan, math.Inf(1)), 1),
		math.IsInf(math.Min(math.Inf(-1), nan), -1), math.IsNaN(math.Max(nan, 1))
}
`

var (
	nan  = math.NaN()
	inf  = math.Inf(1)
	ninf = math.Inf(-1)
	nz   = math.Copysign(0, -1)
)

type minMaxRow struct {
	fn   string
	a, b float64
	want float64 // NaN for any NaN, whose sign and payload are not fixed
}

var minMaxRows = []minMaxRow{
	// the rows of the contract, also with the arguments swapped
	{"Max", nan, inf, inf}, {"Max", inf, nan, inf}, {"Min", nan, ninf, ninf}, {"Min", ninf, nan, ninf},
	{"Max", nan, 1, nan}, {"Max", 1, nan, nan}, {"Min", nz, 0, nz}, {"Min", 0, nz, nz}, {"Max", nz, 0, 0}, {"Max", 0, nz, 0},
	// otherwise NaN wins, also over the other infinity
	{"Min", nan, 1, nan}, {"Min", 1, nan, nan}, {"Max", nan, ninf, nan}, {"Max", ninf, nan, nan},
	{"Min", nan, inf, nan}, {"Min", inf, nan, nan}, {"Max", nan, nan, nan}, {"Min", nan, nan, nan},
	{"Max", nan, nz, nan}, {"Min", 0, nan, nan},
	// the zeros and the ordinary cases
	{"Max", nz, nz, nz}, {"Min", nz, nz, nz}, {"Max", 0, 0, 0}, {"Min", 0, 0, 0},
	{"Max", 1, 2, 2}, {"Max", 2, 1, 2}, {"Min", 1, 2, 1}, {"Min", -1, -2, -2}, {"Max", ninf, 5, 5}, {"Min", inf, 5, 5},
	{"Max", inf, ninf, inf}, {"Min", inf, ninf, ninf}, {"Max", ninf, ninf, ninf}, {"Min", inf, inf, inf},
	{"Max", 5e-324, nz, 5e-324}, {"Min", -5e-324, 0, -5e-324}, {"Max", math.MaxFloat64, inf, inf},
	// the builtins give NaN for any NaN
	{"BuiltinMax", nan, inf, nan}, {"BuiltinMax", inf, nan, nan}, {"BuiltinMin", nan, ninf, nan},
	{"BuiltinMin", ninf, nan, nan}, {"BuiltinMin", nz, 0, nz}, {"BuiltinMax", nz, 0, 0}, {"BuiltinMax", 1, 2, 2},
	{"BuiltinMin", inf, 5, 5},
}

// native are the functions of Go, which the rows call.
var minMaxNative = map[string]func(a, b float64) float64{
	"Max": math.Max, "Min": math.Min,
	"BuiltinMax": func(a, b float64) float64 { return max(a, b) },
	"BuiltinMin": func(a, b float64) float64 { return min(a, b) },
}

// sameFloat tells, whether got is want by its bits, or any NaN, if want is NaN.
func sameFloat(got, want float64) bool {
	if math.IsNaN(want) {
		return math.IsNaN(got)
	}
	return math.Float64bits(got) == math.Float64bits(want)
}

// TestMinMaxBits compares math.Min and math.Max natively, in goja and in the vm with the contract by their bits.
func TestMinMaxBits(t *testing.T) {
	engines := loadTextNum(t, minMaxSrc)
	for _, r := range minMaxRows {
		t.Run(fmt.Sprintf("%s(%v,%v)", r.fn, r.a, r.b), func(t *testing.T) {
			if got := minMaxNative[r.fn](r.a, r.b); !sameFloat(got, r.want) {
				t.Errorf("native: %s(%v, %v) = %v (%#016x), want %v", r.fn, r.a, r.b, got, math.Float64bits(got), r.want)
			}
			for _, e := range engines {
				res := e.call(t, r.fn, r.a, r.b)
				got, sign := res[0].(float64), res[1].(bool)
				if !sameFloat(got, r.want) || (!math.IsNaN(r.want) && sign != math.Signbit(r.want)) {
					t.Errorf("%s: %s(%v, %v) = %v (%#016x, sign bit %v), want %v", e.name, r.fn, r.a, r.b, got,
						math.Float64bits(got), sign, r.want)
				}
			}
		})
	}

	// the same with the values made in the program
	for _, e := range engines {
		res := e.call(t, "Zeros")
		if want := []any{false, true, true, true, true}; fmt.Sprint(res) != fmt.Sprint(want) {
			t.Errorf("%s: Signbit(Max(-0, 0)), Signbit(Min(0, -0)), IsInf(Max(NaN, +Inf)), IsInf(Min(-Inf, NaN)), IsNaN(Max(NaN, 1)) = %v, want %v",
				e.name, res, want)
		}
	}
}
