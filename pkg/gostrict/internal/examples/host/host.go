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
	"go.wdy.de/nago/pkg/gostrict/std/js"
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

// Tabulate returns f of each index from 0 to n, which shows a callback, which the host calls during the call.
//
//gostrict:js (n, f) => { const out = []; for (let i = 0; i < n; i++) out.push(f(i)); return out; }
func Tabulate(n int32, f func(i int32) geom.Rect) list.List[geom.Rect] {
	res := list.New[geom.Rect]()
	for i := range n {
		res.Append(f(i))
	}
	return res
}

// Fold folds the rectangles with f and stops at its first error, which shows a callback with several results.
//
//gostrict:js (rs, f) => { let acc = 0; for (const r of rs) { const [v, err] = f(acc, r); if (err !== null) return [acc, err]; acc = v; } return [acc, null]; }
func Fold(rs list.List[geom.Rect], f func(acc float64, r geom.Rect) (float64, error)) (float64, error) {
	acc := 0.0
	for r := range rs.Values() {
		v, err := f(acc, r)
		if err != nil {
			return acc, err
		}
		acc = v
	}
	return acc, nil
}

// Later is a function of the program, which the host keeps and calls after the call, like a window, which renders
// a table, which a program has created.
type Later struct{ ref js.Ref }

// Keep keeps f.
//
//gostrict:js (f) => ({ f })
func Keep(f func(x float64) float64) Later {
	return Later{ref: js.NewRef(f)}
}
