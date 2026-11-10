// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package longs computes with int64 and uint64, which wrap around in every engine like in Go.
package longs

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Long has fields, whose zero values are 64 bit integers.
type Long struct {
	A int64
	B uint64
}

func Wrap(a, b int64) (int64, int64, int64, int64) {
	return a + b, a * b, a - b, -a
}

func Div(a, b int64) (int64, int64, bool) {
	if b == 0 {
		return 0, 0, false
	}
	return a / b, a % b, true
}

func DivU(a, b uint64) uint64 {
	d := b
	if d > 0 {
		return a/d + a%d
	}
	return 0
}

func Bits(x uint64) (uint64, uint64, int64, uint64) {
	rot := (x << 7) | (x >> 57)
	y := int64(x)
	x2 := x
	x2 &^= 0xff
	return rot ^ (x &^ 0xff), ^x, y >> 3, x2
}

func Shifts(a int32, b uint32, c int64, d uint64, n uint32) (int32, int32, uint32, uint32, int64, int64, uint64, uint64) {
	return a << n, a >> n, b << n, b >> n, c << n, c >> n, d << uint64(n), d >> uint64(n)
}

func Conv(a int32, b uint32, c int64, d uint64) (int64, uint64, int64, uint64, int32, uint32, int32, uint32, int64, uint64, float64, float64) {
	return int64(a), uint64(a), int64(b), uint64(b), int32(c), uint32(c), int32(d), uint32(d), int64(d), uint64(c), float64(c), float64(d)
}

func Floats(f float64) (int64, uint64, int64) {
	return num.Trunc64(f), num.TruncU64(f), num.Round64(f)
}

func Text(a int64, b uint64) (string, string) {
	return text.FromInt64(a), text.FromUint64(b)
}

func Parse(s string) (int64, bool, uint64, bool) {
	_, err := text.ParseInt64(s)
	_, err2 := text.ParseUint64(s)
	return opt.OrZeroErr(text.ParseInt64(s)), err == nil, opt.OrZeroErr(text.ParseUint64(s)), err2 == nil
}

func Ranges(n int64, m uint64) (int64, uint64, int64) {
	var sum int64
	for i := range n {
		sum += i * 3
	}
	var usum uint64
	for j := range m {
		usum += j
	}
	k := int64(10)
	for c := int64(0); c < n; c++ {
		k--
	}
	switch k {
	case 7:
		k = 700
	case 5:
		k = 500
	}
	return sum, usum, k
}

func Containers(a int64, b uint64) (int64, uint64, int64, int32, int64, uint64, bool) {
	l := list.Of(a, a+1)
	first := l.GetOr(int32(0), 0)
	m := dict.New[uint64, int64]()
	m.Set(b, a)
	m.Set(b, a+2)
	v := m.GetOr(b, 0)
	var s Long
	t := Long{A: a}
	u := t
	u.B = b
	zero := list.Make[int64](1)
	z := opt.OrZero(zero.Get(int32(0)))
	return first, s.B, v, m.Len(), z + s.A, u.B, t == Long{A: a}
}

func MinMax(a, b int64, c, d uint64) (int64, int64, uint64, uint64) {
	return min(a, b), max(a, b), min(c, d), max(c, d)
}
