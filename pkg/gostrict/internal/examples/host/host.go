// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package host is a header package: its functions run natively in Go and through their bindings in JavaScript.
//
//gostrict:header
package host

import (
	"errors"
	"strconv"

	"go.wdy.de/nago/pkg/gostrict/internal/examples/geom"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

//gostrict:js (a, b) => a * b + 1
func Mix(a, b float64) float64 {
	return a*b + 1
}

//gostrict:js (x, lo, hi) => Math.min(hi, Math.max(lo, x))
func Clamp(x, lo, hi float64) float64 {
	return min(hi, max(lo, x))
}

//gostrict:js (n) => "#" + n
func Label(n int32) string {
	return "#" + strconv.Itoa(int(n))
}

//gostrict:js (a, b) => b === 0 ? [0, $errors_New("division by zero")] : [a / b, null]
func Ratio(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// Grow enlarges the rectangle by d on each side, which shows a record as a parameter and a result.
//
//gostrict:js (r, d) => ({ X: r.X - d, Y: r.Y - d, W: r.W + 2 * d, H: r.H + 2 * d })
func Grow(r geom.Rect, d float64) geom.Rect {
	return geom.Rect{X: r.X - d, Y: r.Y - d, W: r.W + 2*d, H: r.H + 2*d}
}

// Summarize returns the statistics of the rectangles, which shows lists in records.
//
//gostrict:js (rs, label) => ({ Count: rs.length, Sum: rs.reduce((s, r) => s + r.W * r.H, 0), Label: label, Parts: rs.filter((r) => r.W > 0) })
func Summarize(rs list.List[geom.Rect], label string) geom.Stats {
	st := geom.Stats{Label: label, Parts: list.New[geom.Rect]()}
	for _, r := range rs.All() {
		st.Count++
		st.Sum += r.W * r.H
		if r.W > 0 {
			st.Parts.Append(r)
		}
	}
	return st
}
