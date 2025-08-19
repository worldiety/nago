// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package barchart

import (
	"fmt"
	"log/slog"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/chart"
)

// TBarChart is a data visualization component (Bar Chart).
// It represents categorical data with rectangular bars, supporting
// horizontal or vertical orientation, stacked bars, and markers.
// The chart can be customized by providing multiple data series
// and additional visual indicators (markers).
type TBarChart struct {
	chart      chart.Chart
	series     []chart.Series
	markers    []Marker
	horizontal bool
	stacked    bool
}

// BarChart creates a new bar chart with default (vertical, non-stacked) configuration.
func BarChart(chart chart.Chart) TBarChart {
	return TBarChart{
		chart:      chart,
		horizontal: false,
		stacked:    false,
	}
}

// Chart sets the underlying chart configuration for the bar chart.
func (c TBarChart) Chart(chart chart.Chart) TBarChart {
	c.chart = chart
	return c
}

// Series defines the data series to be displayed in the bar chart.
func (c TBarChart) Series(series []chart.Series) TBarChart {
	c.series = series
	return c
}

// Markers adds markers to the bar chart to highlight values or ranges.
func (c TBarChart) Markers(markers []Marker) TBarChart {
	c.markers = markers
	return c
}

// Horizontal sets whether the bar chart is rendered horizontally.
func (c TBarChart) Horizontal(horizontal bool) TBarChart {
	c.horizontal = horizontal
	return c
}

// Stacked sets whether multiple series are stacked instead of grouped.
func (c TBarChart) Stacked(stacked bool) TBarChart {
	c.stacked = stacked
	return c
}

// Render converts the TBarChart into its renderable proto.BarChart representation.
// It maps the configured chart, series, and optional markers into the underlying
// protocol buffer format, which is used by the UI system to display the chart.
func (c TBarChart) Render(ctx core.RenderContext) core.RenderNode {
	protoSeries := make([]proto.ChartSeries, len(c.series))
	var protoMarkers []proto.BarChartMarker

	for i, series := range c.series {
		protoSeries[i] = series.Ora()
	}

	if !c.stacked {
		protoMarkers = make([]proto.BarChartMarker, len(c.markers))

		for k, marker := range c.markers {
			protoMarkers[k] = marker.Ora()
		}
	} else if len(c.markers) > 0 {
		slog.Warn(fmt.Sprintf("BarChart: markers are not supported for stacked charts"))
	}

	return &proto.BarChart{
		Chart:      c.chart.Ora(),
		Series:     protoSeries,
		Markers:    protoMarkers,
		Horizontal: proto.Bool(c.horizontal),
		Stacked:    proto.Bool(c.stacked),
	}
}
