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
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/chart"
	"go.wdy.de/nago/presentation/ui/linechart"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_66")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			var lineChartSeries []chart.Series
			var lineChartSeries2 []chart.Series

			grow := core.AutoState[float64](wnd)

			lineChartSeries = append(lineChartSeries, chart.Series{
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
					{
						X: "2011",
						Y: 450,
					},
					{
						X: "2012",
						Y: 600,
					},
					{
						X: "2013",
						Y: grow.Get(),
					},
				},
			})
			lineChartSeries = append(lineChartSeries, chart.Series{
				Label: "series 2",
				DataPoints: []chart.DataPoint{
					{
						X: "2009",
						Y: 400,
					},
					{
						X: "2010",
						Y: 500,
					},
					{
						X: "2011",
						Y: 500,
					},
					{
						X: "2012",
						Y: 50,
					},
					{
						X: "2013",
						Y: grow.Get(),
					},
				},
			})
			lineChartSeries2 = append(lineChartSeries2, chart.Series{
				Label: "series 1",
				Type:  chart.ChartSeriesTypeColumn,
				DataPoints: []chart.DataPoint{
					{
						X: "2009",
						Y: 200,
					},
					{
						X: "2010",
						Y: 300,
					},
					{
						X: "2011",
						Y: 450,
					},
					{
						X: "2012",
						Y: 600,
					},
					{
						X: "2013",
						Y: grow.Get(),
					},
				},
			})
			lineChartSeries2 = append(lineChartSeries2, chart.Series{
				Label: "series 2",
				Type:  chart.ChartSeriesTypeArea,
				DataPoints: []chart.DataPoint{
					{
						X: "2009",
						Y: 400,
					},
					{
						X: "2010",
						Y: 500,
					},
					{
						X: "2011",
						Y: 500,
					},
					{
						X: "2012",
						Y: 50,
					},
					{
						X: "2013",
						Y: grow.Get(),
					},
				},
			})

			markers := linechart.Markers{
				Size: 5,
			}

			colorPalette := []ui.Color{
				ui.M0,
				ui.A0,
			}
			colorPalette2 := []ui.Color{
				ui.M0,
				ui.SE0,
			}

			chart1 := chart.Chart{
				Colors:       colorPalette,
				Frame:        ui.Frame{}.Size(ui.L320, ui.L200),
				Downloadable: false,
			}
			chart2 := chart.Chart{
				Colors: colorPalette2,
				Frame:  ui.Frame{}.Size(ui.L320, ui.L200),
			}
			chart3 := chart.Chart{
				Frame:         ui.Frame{}.Size(ui.L320, ui.L200),
				NoDataMessage: "Ein LineChart ohne Daten",
			}
			return ui.VStack(
				ui.Text("line chart demo"),
				breadcrumb.Breadcrumbs().
					Item("home", nil).
					Item("a", nil).
					Item("b", nil),
				linechart.LineChart(chart1).Series(lineChartSeries).Markers(markers),
				linechart.LineChart(chart1).Series(lineChartSeries).Curve(linechart.CurveSmooth),
				linechart.LineChart(chart1).Series(lineChartSeries).Curve(linechart.CurveStepline),
				linechart.LineChart(chart2).Series(lineChartSeries2),
				linechart.LineChart(chart3).Series([]chart.Series{}),
				ui.PrimaryButton(func() {
					grow.Set(grow.Get() + 100)
				}).Title("grow 2013"),
			)
		})
	}).Run()
}
