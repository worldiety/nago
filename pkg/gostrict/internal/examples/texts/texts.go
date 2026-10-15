// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package texts processes strings with umlauts and emojis, which Go and JavaScript encode differently.
package texts

import (
	"strings"

	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

func Reverse(s string) string {
	chars := text.Chars(s)
	r := list.New[string]()
	for i := chars.Len() - 1; i >= 0; i-- {
		if c, ok := chars.Get(i); ok {
			r.Append(c)
		}
	}
	return text.Join(r, "")
}

func Stats(s string) (int32, int32, int32) {
	return text.RuneCount(s), text.ByteLen(s), text.Index(s, "ö")
}

func Words(s string) string {
	counts := dict.New[string, int32]()
	for _, w := range text.Split(strings.TrimSpace(s), " ").All() {
		if w == "" {
			continue
		}
		n, _ := counts.Get(w)
		counts.Set(w, n+1)
	}
	out := ""
	for w, n := range counts.All() {
		out += w + "=" + text.FromInt(n) + ";"
	}
	return out
}

// Order compares across the range, where UTF-16 sorts differently than code points.
func Order(a, b string) (bool, int32) {
	return a < b, text.Compare(a, b)
}

func Parse(s string) string {
	n, err := text.ParseInt(s)
	if err != nil {
		return "error: " + err.Error()
	}
	return "ok: " + text.FromInt(n*2)
}

func Misc(s string) string {
	return strings.ReplaceAll(s, "", "-") + "|" + text.Repeat(s, 2) + "|" + text.UpperASCII(s) + "|" + text.FromInt(text.Count(s, "a")) + "|" + strings.TrimSpace(s)
}
