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
	"go.wdy.de/nago/presentation/ui/barchart"
	"go.wdy.de/nago/presentation/ui/chart"
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
			var barChartSeries []chart.Series

			barChartSeries = append(barChartSeries, chart.Series{
				Label: "series 1",
				DataPoints: []chart.DataPoint{
					{
						X: "2009",
						Y: 200,
					},
					{
						X: "2010",
						Y: 300,
					},
				},
			})
			barChartSeries = append(barChartSeries, chart.Series{
				Label: "series 2",
				DataPoints: []chart.DataPoint{
					{
						X: "2009",
						Y: 400,
					},
					{
						X: "2011",
						Y: 500,
					},
				},
			})

			markers := []barchart.Marker{
				{
					Label:          "marker 1",
					SeriesIndex:    0,
					DataPointIndex: 0,
					Value:          "180",
					Color:          ui.SE0,
					Width:          10,
					Height:         10,
					Round:          true,
				},
				{
					Label:          "marker 2",
					SeriesIndex:    1,
					DataPointIndex: 0,
					Value:          "450",
					Color:          ui.SE0,
				},
				{
					Label:          "marker 3",
					SeriesIndex:    1,
					DataPointIndex: 1,
					Value:          "450",
					Color:          ui.SE0,
					Dashed:         true,
				},
			}

			colorPalette := []ui.Color{
				ui.M0,
				ui.A0,
			}

			chart1 := chart.Chart{
				Labels:       []string{"2009", "2010", "2011", "2012"},
				Colors:       colorPalette,
				Frame:        ui.Frame{}.Size(ui.L320, ui.L200),
				Downloadable: false,
			}
			chart2 := chart.Chart{
				Colors: colorPalette,
				Frame:  ui.Frame{}.Size(ui.L320, ui.L200),
			}
			chart3 := chart.Chart{
				Labels:     []string{"2009", "2010", "2011", "2012"},
				Colors:     colorPalette,
				Frame:      ui.Frame{}.Size(ui.L320, ui.L200),
				XAxisTitle: "x-axis",
				YAxisTitle: "y-axis",
			}
			chart4 := chart.Chart{
				Frame:         ui.Frame{}.Size(ui.L320, ui.L200),
				NoDataMessage: "Ein BarChart ohne Daten",
			}

			return ui.VStack(
				ui.Text("bar chart demo"),
				barchart.BarChart(chart1).Series(barChartSeries).Markers(markers),
				barchart.BarChart(chart2).Series(barChartSeries).Markers(markers).Horizontal(true),
				barchart.BarChart(chart3).Series(barChartSeries).Markers(markers).Stacked(true),
				barchart.BarChart(chart4).Series([]chart.Series{}),
			)
		})
	}).Run()
}
