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

type TLineChart struct {
	chart   chart.Chart
	series  []chart.Series
	markers Markers
	curve   Curve
}

func LineChart(chart chart.Chart) TLineChart {
	return TLineChart{
		chart: chart,
	}
}

func (c TLineChart) Chart(chart chart.Chart) TLineChart {
	c.chart = chart
	return c
}

func (c TLineChart) Series(series []chart.Series) TLineChart {
	c.series = series
	return c
}

func (c TLineChart) Markers(markers Markers) TLineChart {
	c.markers = markers
	return c
}

func (c TLineChart) Curve(curve Curve) TLineChart {
	c.curve = curve
	return c
}

func (c TLineChart) Render(ctx core.RenderContext) core.RenderNode {
	protoSeries := make([]proto.ChartSeries, len(c.series))

	for i, series := range c.series {
		protoSeries[i] = series.Ora()
	}

	return &proto.LineChart{
		Chart:  c.chart.Ora(),
		Series: protoSeries,
		Curve:  c.curve.ora(),
	}
}
