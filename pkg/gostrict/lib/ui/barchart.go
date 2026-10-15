// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/ui/barchart"
)

// TBarChart is a bar chart, which shows categorical data with rectangular bars, see barchart.TBarChart of nago.
type TBarChart struct{ ref js.Ref }

func (TBarChart) isView() {}

// Horizontal sets whether the bar chart is rendered horizontally.
//
//gostrict:host
func (c TBarChart) Horizontal(horizontal bool) TBarChart {
	return c.with(func(v barchart.TBarChart) barchart.TBarChart { return v.Horizontal(horizontal) })
}

// Stacked sets whether multiple series are stacked instead of grouped.
//
//gostrict:host
func (c TBarChart) Stacked(stacked bool) TBarChart {
	return c.with(func(v barchart.TBarChart) barchart.TBarChart { return v.Stacked(stacked) })
}

// Markers adds markers to the bar chart to highlight values or ranges.
//
//gostrict:host
func (c TBarChart) Markers(markers ...BarChartMarker) TBarChart {
	return c.with(func(v barchart.TBarChart) barchart.TBarChart { return v.Markers(c2BarChartMarkers(markers)) })
}
