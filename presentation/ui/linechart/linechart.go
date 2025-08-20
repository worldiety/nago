// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package linechart

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/chart"
)

// TLineChart is a composite component (Line Chart).
// This component renders a line chart based on provided data series.
// It supports different curve styles (straight, smooth, stepline) and
// can display markers for data points or thresholds.
// Typical usage is in dashboards, analytics tools, or reporting interfaces.
type TLineChart struct {
	chart   chart.Chart    // underlying chart configuration and layout
	series  []chart.Series // data series that define the plotted lines
	markers Markers        // optional markers (e.g. points, thresholds, highlights)
	curve   Curve          // curve style for connecting data points
}

// LineChart returns a TLineChart initialized with the given chart.
func LineChart(chart chart.Chart) TLineChart {
	return TLineChart{
		chart: chart,
	}
}

// Chart sets the chart config.
func (c TLineChart) Chart(chart chart.Chart) TLineChart {
	c.chart = chart
	return c
}

// Series sets the data series.
func (c TLineChart) Series(series []chart.Series) TLineChart {
	c.series = series
	return c
}

// Markers sets the markers config.
func (c TLineChart) Markers(markers Markers) TLineChart {
	c.markers = markers
	return c
}

// Curve sets the curve style.
func (c TLineChart) Curve(curve Curve) TLineChart {
	c.curve = curve
	return c
}

// Render builds the proto.LineChart node.
func (c TLineChart) Render(ctx core.RenderContext) core.RenderNode {
	protoSeries := make([]proto.ChartSeries, len(c.series))

	for i, series := range c.series {
		protoSeries[i] = series.Ora()
	}

	return &proto.LineChart{
		Chart:   c.chart.Ora(),
		Series:  protoSeries,
		Curve:   c.curve.ora(),
		Markers: c.markers.Ora(),
	}
}
