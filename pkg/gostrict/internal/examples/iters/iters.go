// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package iters ranges over iterators, whose bodies are translated into callbacks.
package iters

import (
	"iter"

	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

func Evens(n int32) iter.Seq[int32] {
	return func(yield func(int32) bool) {
		for i := int32(0); i < n; i += 2 {
			if !yield(i) {
				return
			}
		}
	}
}

func Indexed(s string) iter.Seq2[int32, string] {
	return func(yield func(int32, string) bool) {
		for i, c := range text.Chars(s).All() {
			if !yield(i, c) {
				return
			}
		}
	}
}

func FirstOver(n, limit int32) int32 {
	for v := range Evens(n) {
		if v > limit {
			return v
		}
	}
	return -1
}

func Sum(n int32) int32 {
	s := int32(0)
	for v := range Evens(n) {
		if v == 4 {
			continue
		}
		if v > 10 {
			break
		}
		s += v
	}
	return s
}

func Nested(n int32) (int32, string) {
	for a := range Evens(n) {
		for b := range Evens(n) {
			if a+b == 6 && a > 0 {
				return a*10 + b, "found"
			}
		}
	}
	return -1, "none"
}

func Switch(s string) string {
	out := ""
outer:
	for i, c := range text.Chars(s).All() {
		switch c {
		case "a", "e":
			out += "V"
		case "x":
			break outer
		case "-":
			continue
		default:
			if i > 5 {
				break
			}
			out += c
		}
		out += "."
	}
	return out
}

func Live() string {
	m := dict.New[string, int32]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	out := ""
	for k, v := range m.All() {
		if k == "a" {
			m.Delete("b")
			m.Set("d", 4)
		}
		out += k + text.FromInt(v)
	}
	return out + "|" + text.FromInt(m.Len())
}
