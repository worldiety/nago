// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package money computes exactly with big.Int and decimal.Dec, which must agree in Go and JavaScript to the last
// digit.
package money

import (
	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/decimal"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Ints combines two integers of any size, which cross the boundary as big.Int.
func Ints(a, b big.Int) (big.Int, big.Int, big.Int, big.Int, big.Int, bool, int32) {
	q, r, ok := big.QuoRem(a, b)
	if !ok {
		return big.Add(a, b), big.Sub(a, b), big.Mul(a, b), big.Zero(), big.Zero(), false, big.Cmp(a, b)
	}
	return big.Add(a, b), big.Sub(a, b), big.Mul(a, b), q, r, true, big.Cmp(a, b)
}

// Factorial returns n! in decimal digits, which exceeds the exact integers of float64 from 19 on.
func Factorial(n int32) string {
	f := big.FromInt32(1)
	for i := range n {
		f = big.Mul(f, big.FromInt32(i+1))
	}
	return big.String(f)
}

// Parse parses s as big.Int and decimal.Dec and formats both again.
func Parse(s string) (string, string) {
	i := "invalid"
	if n, ok := big.Parse(s); ok {
		i = big.String(n)
	}
	d := "invalid"
	if n, ok := decimal.Parse(s); ok {
		d = decimal.String(n)
	}
	return i, d
}

// Calc combines two decimals and divides them to the scale s.
func Calc(a, b string, s int32, mode int32) (string, string, string, string, bool, int32) {
	x, okA := decimal.Parse(a)
	y, okB := decimal.Parse(b)
	if !okA || !okB {
		return "", "", "", "", false, 0
	}
	q, ok := decimal.Quo(x, y, s, rounding(mode))
	return decimal.String(decimal.Add(x, y)), decimal.String(decimal.Sub(x, y)), decimal.String(decimal.Mul(x, y)), decimal.String(q), ok, decimal.Cmp(x, y)
}

// Round rounds each of the amounts to the scale s by each mode.
func Round(amounts list.List[string], s int32) list.List[string] {
	res := list.New[string]()
	for a := range amounts.Values() {
		d, ok := decimal.Parse(a)
		if !ok {
			res.Append("invalid " + a)
			continue
		}
		for m := range int32(6) {
			res.Append(decimal.String(decimal.Round(d, s, rounding(m))))
		}
	}
	return res
}

// Invoice sums the net amounts of the lines, rounds the tax at the rate in percent commercially to cents and
// returns net, tax and gross.
func Invoice(lines list.List[string], rate string) (string, string, string, bool) {
	net := decimal.Zero()
	for l := range lines.Values() {
		d, ok := decimal.Parse(l)
		if !ok {
			return "", "", "", false
		}
		net = decimal.Add(net, d)
	}
	r, ok := decimal.Parse(rate)
	if !ok {
		return "", "", "", false
	}
	tax, _ := decimal.Quo(decimal.Mul(net, r), decimal.FromInt32(100), 2, decimal.HalfUp)
	net = decimal.Round(net, 2, decimal.HalfUp)
	return decimal.String(net), decimal.String(tax), decimal.String(decimal.Add(net, tax)), true
}

func rounding(m int32) decimal.Rounding {
	switch m {
	case 1:
		return decimal.HalfEven
	case 2:
		return decimal.Down
	case 3:
		return decimal.Up
	case 4:
		return decimal.Floor
	case 5:
		return decimal.Ceiling
	}
	return decimal.HalfUp
}

// Limits clamps the scales of literals out of range, rounds a product beyond decimal.MaxScale and converts a.
func Limits(a big.Int, n uint32, e int32) (string, string, int32, int32, bool, float64, big.Int, big.Int) {
	negative := decimal.String(decimal.Dec{M: a, S: -3})
	huge := decimal.Dec{M: a, S: 5000}
	// 10^-999 * 0.15 has the scale 1001 and rounds to 2 * 10^-1000
	tiny := decimal.Mul(decimal.New(big.FromInt32(1), 999), decimal.New(big.FromInt32(15), 2))
	_, long := decimal.Parse("0." + text.Repeat("1", decimal.MaxScale+1))
	i := int32(-1)
	if v, ok := big.Int32(a); ok {
		i = v
	}
	return negative, decimal.String(decimal.Round(huge, 3, decimal.HalfUp)), decimal.Cmp(tiny, decimal.New(big.FromInt32(2), 1000)), i, long, big.Float64(a), big.FromUint32(n), big.Pow10(e)
}
