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
	"go.wdy.de/nago/presentation/ui/chart"
	"go.wdy.de/nago/presentation/ui/svgchart"
)

func init() {
	app.Register("svg-bar-chart", func(wnd core.Window) core.View {
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

		return svgchart.BarChart(chart.Chart{
			Frame:      Frame{}.Size(L560, L320),
			XAxisTitle: "Quarter",
			YAxisTitle: "Orders",
		}).Series(series)
	})

	app.Register("svg-line-chart", func(wnd core.Window) core.View {
		series := []chart.Series{
			{
				Label: "Visitors",
				DataPoints: []chart.DataPoint{
					{X: "Mon", Y: 320}, {X: "Tue", Y: 410}, {X: "Wed", Y: 380},
					{X: "Thu", Y: 520}, {X: "Fri", Y: 610}, {X: "Sat", Y: 290}, {X: "Sun", Y: 240},
				},
			},
			{
				Label: "Sign-ups",
				DataPoints: []chart.DataPoint{
					{X: "Mon", Y: 40}, {X: "Tue", Y: 65}, {X: "Wed", Y: 52},
					{X: "Thu", Y: 90}, {X: "Fri", Y: 110}, {X: "Sat", Y: 35}, {X: "Sun", Y: 28},
				},
			},
		}

		return svgchart.LineChart(chart.Chart{
			Frame: Frame{}.Size(L560, L320),
		}).Series(series).Curve(svgchart.CurveSmooth).Markers(svgchart.Markers{Size: 4})
	})

	app.Register("svg-pie-chart", func(wnd core.Window) core.View {
		series := []chart.Series{
			{
				Label: "Browser",
				DataPoints: []chart.DataPoint{
					{X: "Chrome", Y: 64}, {X: "Safari", Y: 19}, {X: "Firefox", Y: 9}, {X: "Other", Y: 8},
				},
			},
		}

		cfg := chart.Chart{Frame: Frame{}.Size(L320, L256)}

		return HStack(
			svgchart.PieChart(cfg).Series(series),
			svgchart.PieChart(cfg).Series(series).ShowAsDonut(true),
		).Gap(L32)
	})
}
