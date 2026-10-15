// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package hostuse calls the functions of a header package.
package hostuse

import (
	"go.wdy.de/nago/pkg/gostrict/internal/examples/geom"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/host"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

func Run(a, b float64) (string, float64) {
	r, err := host.Ratio(a, b)
	if err != nil {
		return host.Label(-1) + " " + err.Error(), 0
	}
	return host.Label(7), host.Clamp(host.Mix(a, b)+r, 0, 100)
}

// Records passes records through the functions of the header package and returns a summary.
func Records(d float64) (string, float64, int32) {
	r := host.Grow(geom.Rect{X: 1, Y: 2, W: 3, H: 4}, d)
	st := host.Summarize(list.Of(r, geom.Rect{W: 0, H: 9}, geom.Rect{W: 2, H: 2}), "rects")
	first, ok := st.Parts.Get(0)
	if !ok {
		return st.Label, st.Sum, st.Count
	}
	return st.Label + " " + text.FromFloat(first.X, 2), st.Sum, st.Parts.Len()
}
