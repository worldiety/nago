// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package barchart

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/chart"
	"log/slog"
)

type TBarChart struct {
	chart      chart.Chart
	series     []chart.Series
	markers    []Marker
	horizontal bool
	stacked    bool
}

func BarChart(chart chart.Chart) TBarChart {
	return TBarChart{
		chart:      chart,
		horizontal: false,
		stacked:    false,
	}
}

func (c TBarChart) Chart(chart chart.Chart) TBarChart {
	c.chart = chart
	return c
}

func (c TBarChart) Series(series []chart.Series) TBarChart {
	c.series = series
	return c
}

func (c TBarChart) Markers(markers []Marker) TBarChart {
	c.markers = markers
	return c
}

func (c TBarChart) Horizontal(horizontal bool) TBarChart {
	c.horizontal = horizontal
	return c
}

func (c TBarChart) Stacked(stacked bool) TBarChart {
	c.stacked = stacked
	return c
}

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
