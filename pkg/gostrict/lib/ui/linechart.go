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
	"go.wdy.de/nago/presentation/ui/linechart"
)

// TLineChart is a line chart, which plots data series as lines, see linechart.TLineChart of nago.
type TLineChart struct{ ref js.Ref }

func (TLineChart) isView() {}

// Markers sets the markers of the data points.
//
//gostrict:host
func (c TLineChart) Markers(markers LineChartMarkers) TLineChart {
	return c.with(func(v linechart.TLineChart) linechart.TLineChart { return v.Markers(markers.nago()) })
}

// Curve sets the curve style, which connects the data points.
//
//gostrict:host
func (c TLineChart) Curve(curve LineChartCurve) TLineChart {
	return c.with(func(v linechart.TLineChart) linechart.TLineChart { return v.Curve(linechart.Curve(curve)) })
}
