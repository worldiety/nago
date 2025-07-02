// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/charts"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_65")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			var barChartSeries []charts.TBarChartSeries

			barChartSeries = append(barChartSeries, charts.TBarChartSeries{
				Label: "series 1",
				DataPoints: []charts.TBarChartDataPoint{
					{
						X: "2009",
						Y: "200",
						Markers: []charts.TBarChartMarker{
							{
								Label:   "marker 1",
								Value:   "180",
								Color:   ui.SE0,
								Width:   10,
								Height:  10,
								IsRound: true,
							},
						},
					},
					{
						X: "2010",
						Y: "300",
					},
				},
			})

			barChartSeries = append(barChartSeries, charts.TBarChartSeries{
				Label: "series 2",
				DataPoints: []charts.TBarChartDataPoint{
					{
						X: "2009",
						Y: "400",
						Markers: []charts.TBarChartMarker{
							{
								Label: "marker 2",
								Value: "450",
								Color: ui.SE0,
							},
						},
					},
					{
						X: "2010",
						Y: "500",
						Markers: []charts.TBarChartMarker{
							{
								Label:    "marker 3",
								Value:    "450",
								Color:    ui.SE0,
								IsDashed: true,
							},
						},
					},
				},
			})

			colorPalette := []ui.Color{
				ui.M0,
				ui.A0,
			}

			return ui.VStack(
				ui.Text("bar chart demo"),
				charts.BarChart().Series(barChartSeries).Colors(colorPalette).Frame(ui.Frame{}.Size(ui.L320, ui.L200)),
				charts.BarChart().Series(barChartSeries).Colors(colorPalette).IsHorizontal(true).Frame(ui.Frame{}.Size(ui.L320, ui.L200)),
				charts.BarChart().Series(barChartSeries).Colors(colorPalette).IsStacked(true).Frame(ui.Frame{}.Size(ui.L320, ui.L200)),
				charts.BarChart().Series([]charts.TBarChartSeries{}).NoDataMessage("Ein BarChart ohne Daten").Frame(ui.Frame{}.Size(ui.L320, ui.L200)),
			)
		})
	}).Run()
}
