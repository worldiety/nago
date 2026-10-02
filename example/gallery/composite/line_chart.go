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
	"go.wdy.de/nago/presentation/ui/linechart"
)

func init() {
	app.Register("line-chart", func(wnd core.Window) core.View {
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

		return linechart.LineChart(chart.Chart{
			Frame: Frame{}.Size(L560, L320),
		}).Series(series).Curve(linechart.CurveSmooth)
	})
}
