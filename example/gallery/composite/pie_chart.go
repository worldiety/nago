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
	"go.wdy.de/nago/presentation/ui/piechart"
)

func init() {
	app.Register("pie-chart", func(wnd core.Window) core.View {
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
			piechart.PieChart(cfg).Series(series),
			piechart.PieChart(cfg).Series(series).ShowAsDonut(true),
		).Gap(L32)
	})
}
