// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package ints computes with int32 and uint32, which wrap around in Go and JavaScript alike.
package ints

import (
	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Hash is FNV-1a over the code points.
func Hash(s string) uint32 {
	h := uint32(2166136261)
	for _, cp := range text.CodePoints(s).All() {
		h ^= uint32(cp)
		h *= 16777619
	}
	return h
}

func Wrap(a, b int32) (int32, int32, int32, int32) {
	return a + b, a * b, a - b, -a
}

func Div(a, b int32) (int32, int32, bool) {
	if b == 0 {
		return 0, 0, false
	}
	return a / b, a % b, true
}

func DivU(a, b uint32) uint32 {
	d := b
	if d > 0 {
		return a/d + a%d
	}
	return 0
}

func Bits(x uint32) (uint32, uint32, int32) {
	rot := (x << 7) | (x >> 25)
	y := int32(x)
	return rot ^ (x &^ 0xff), ^x, y >> 3
}

func Conv(f float64) (int32, uint32, int32) {
	return num.Trunc32(f), num.TruncU32(f), num.Round32(f)
}

// Xorshift is a deterministic random generator.
func Xorshift(seed uint32, n int32) uint32 {
	x := seed
	for range n {
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
	}
	return x
}
