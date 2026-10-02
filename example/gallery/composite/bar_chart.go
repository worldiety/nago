// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/barchart"
	"go.wdy.de/nago/presentation/ui/chart"
)

func init() {
	app.Register("bar-chart", func(wnd core.Window) core.View {
		series := []chart.Series{
			{
				Label: "Online",
				DataPoints: []chart.DataPoint{
					{X: "Q1", Y: 120}, {X: "Q2", Y: 180}, {X: "Q3", Y: 150}, {X: "Q4", Y: 210},
				},
			},
			{
				Label: "Retail",
				DataPoints: []chart.DataPoint{
					{X: "Q1", Y: 90}, {X: "Q2", Y: 110}, {X: "Q3", Y: 130}, {X: "Q4", Y: 160},
				},
			},
		}

		return barchart.BarChart(chart.Chart{
			Frame:      Frame{}.Size(L560, L320),
			XAxisTitle: "Quarter",
			YAxisTitle: "Orders",
		}).Series(series)
	})
}
