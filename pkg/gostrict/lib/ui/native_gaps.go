// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//go:build !gostrict

package ui

import (
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/barchart"
	"go.wdy.de/nago/presentation/ui/chart"
	"go.wdy.de/nago/presentation/ui/linechart"
	"go.wdy.de/nago/presentation/ui/piechart"
)

func (s IndexState) nago() *core.State[int] {
	return get[*core.State[int]](s.ref)
}

func (s ColorState) nago() *core.State[nui.Color] {
	return get[*core.State[nui.Color]](s.ref)
}

func (c Chart) nago(labels []string) chart.Chart {
	return chart.Chart{
		Labels:        labels,
		Frame:         c.Frame.nago(),
		Downloadable:  c.Downloadable,
		NoDataMessage: c.NoDataMessage,
		XAxisTitle:    c.XAxisTitle,
		YAxisTitle:    c.YAxisTitle,
		LabelRounding: chart.RoundingType(c.LabelRounding),
	}
}

func chartSeriesOf(label string, typ ChartSeriesType, points []ChartDataPoint) chart.Series {
	s := chart.Series{Label: label, Type: chart.ChartSeriesType(typ)}
	for _, p := range points {
		s.DataPoints = append(s.DataPoints, p.nago())
	}
	return s
}

func chartSeriesList(series []TChartSeries) []chart.Series {
	var res []chart.Series
	for _, s := range series {
		res = append(res, get[chart.Series](s.ref))
	}
	return res
}

// chartHolder keeps the chart of nago beside its view, so that a method like Colors changes the chart, which the
// view of nago does not reveal.
type chartHolder[T interface {
	Chart(chart.Chart) T
	Render(core.RenderContext) core.RenderNode
}] struct {
	v T
	c chart.Chart
}

func newChartHolder[T interface {
	Chart(chart.Chart) T
	Render(core.RenderContext) core.RenderNode
}](v T, c chart.Chart) chartHolder[T] {
	return chartHolder[T]{v: v, c: c}
}

func (h chartHolder[T]) Render(ctx core.RenderContext) core.RenderNode {
	return h.v.Render(ctx)
}

func colorsOf(colors []Color) []nui.Color {
	res := make([]nui.Color, 0, len(colors))
	for _, c := range colors {
		res = append(res, nui.Color(c))
	}
	return res
}

func (c TBarChart) with(fn func(v barchart.TBarChart) barchart.TBarChart) TBarChart {
	h := get[chartHolder[barchart.TBarChart]](c.ref)
	h.v = fn(h.v)
	return TBarChart{ref: wrap(h)}
}

func (c TBarChart) withChart(fn func(ch *chart.Chart)) TBarChart {
	h := get[chartHolder[barchart.TBarChart]](c.ref)
	fn(&h.c)
	h.v = h.v.Chart(h.c)
	return TBarChart{ref: wrap(h)}
}

func (c TLineChart) with(fn func(v linechart.TLineChart) linechart.TLineChart) TLineChart {
	h := get[chartHolder[linechart.TLineChart]](c.ref)
	h.v = fn(h.v)
	return TLineChart{ref: wrap(h)}
}

func (c TLineChart) withChart(fn func(ch *chart.Chart)) TLineChart {
	h := get[chartHolder[linechart.TLineChart]](c.ref)
	fn(&h.c)
	h.v = h.v.Chart(h.c)
	return TLineChart{ref: wrap(h)}
}

func (c TPieChart) with(fn func(v piechart.TPieChart) piechart.TPieChart) TPieChart {
	h := get[chartHolder[piechart.TPieChart]](c.ref)
	h.v = fn(h.v)
	return TPieChart{ref: wrap(h)}
}

func (c TPieChart) withChart(fn func(ch *chart.Chart)) TPieChart {
	h := get[chartHolder[piechart.TPieChart]](c.ref)
	fn(&h.c)
	h.v = h.v.Chart(h.c)
	return TPieChart{ref: wrap(h)}
}
