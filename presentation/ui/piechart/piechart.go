// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package piechart

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/chart"
)

type TPieChart struct {
	chart          chart.Chart
	series         []chart.Series
	showAsDonut    bool
	showDataLabels bool
}

func PieChart(chart chart.Chart) TPieChart {
	return TPieChart{
		chart:          chart,
		showDataLabels: true,
	}
}

func (c TPieChart) Chart(chart chart.Chart) TPieChart {
	c.chart = chart
	return c
}

func (c TPieChart) Series(series []chart.Series) TPieChart {
	c.series = series
	return c
}

func (c TPieChart) ShowAsDonut(showAsDonut bool) TPieChart {
	c.showAsDonut = showAsDonut
	return c
}

func (c TPieChart) ShowDataLabels(showDataLabels bool) TPieChart {
	c.showDataLabels = showDataLabels
	return c
}

func (c TPieChart) Render(ctx core.RenderContext) core.RenderNode {
	protoSeries := make([]proto.ChartSeries, len(c.series))

	for i, series := range c.series {
		protoSeries[i] = series.Ora()
	}

	return &proto.PieChart{
		Chart:          c.chart.Ora(),
		Series:         protoSeries,
		ShowAsDonut:    proto.Bool(c.showAsDonut),
		ShowDataLabels: proto.Bool(c.showDataLabels),
	}
}
