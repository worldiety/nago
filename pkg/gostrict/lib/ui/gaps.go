// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/barchart"
	"go.wdy.de/nago/presentation/ui/colorpicker"
	"go.wdy.de/nago/presentation/ui/linechart"
	"go.wdy.de/nago/presentation/ui/pager"
	"go.wdy.de/nago/presentation/ui/piechart"
	"go.wdy.de/nago/presentation/ui/stepper"
	"go.wdy.de/nago/presentation/ui/switcher"
	"go.wdy.de/nago/presentation/ui/tabs"
)

// The functions of this file close the gaps of the mirror, where nago takes a slice, a state of int or a state of
// a color: they take variadic parameters and the states IndexState and ColorState instead.

// IndexState is an index of the window, e.g. of the active tab, which it keeps between two renders.
type IndexState struct{ ref js.Ref }

// ColorState is a color of the window, e.g. of a color picker.
type ColorState struct{ ref js.Ref }

// IndexStateOf returns the state with the id of the window, which is 0 at first, like tabs.NewIndexState of nago.
//
//gostrict:host
func IndexStateOf(wnd Window, id string) IndexState {
	return IndexState{ref: wrap(core.StateOf[int](wnd.nago(), id))}
}

// Get returns the index.
//
//gostrict:host
func (s IndexState) Get() int32 {
	return int32(s.nago().Get())
}

// Set sets the index and renders the window again.
//
//gostrict:host
func (s IndexState) Set(v int32) {
	s.nago().Set(int(v))
}

// Init sets the index of fn, if the state is new.
//
//gostrict:host
func (s IndexState) Init(fn func() int32) IndexState {
	return IndexState{ref: wrap(s.nago().Init(func() int { return int(fn()) }))}
}

// Observe calls fn, when the user changes the index.
//
//gostrict:host
func (s IndexState) Observe(fn func(v int32)) IndexState {
	return IndexState{ref: wrap(s.nago().Observe(func(v int) { fn(int32(v)) }))}
}

// ColorStateOf returns the state with the id of the window, which has no color at first.
//
//gostrict:host
func ColorStateOf(wnd Window, id string) ColorState {
	return ColorState{ref: wrap(core.StateOf[nui.Color](wnd.nago(), id))}
}

// Get returns the color.
//
//gostrict:host
func (s ColorState) Get() Color {
	return Color(s.nago().Get())
}

// Set sets the color and renders the window again.
//
//gostrict:host
func (s ColorState) Set(v Color) {
	s.nago().Set(nui.Color(v))
}

// Init sets the color of fn, if the state is new.
//
//gostrict:host
func (s ColorState) Init(fn func() Color) ColorState {
	return ColorState{ref: wrap(s.nago().Init(func() nui.Color { return nui.Color(fn()) }))}
}

// Observe calls fn, when the user changes the color.
//
//gostrict:host
func (s ColorState) Observe(fn func(v Color)) ColorState {
	return ColorState{ref: wrap(s.nago().Observe(func(v nui.Color) { fn(Color(v)) }))}
}

// InputValue binds the index of the active tab to the state.
//
//gostrict:host
func (c TTabs) InputValue(activeIdx IndexState) TTabs {
	return TTabs{ref: wrap(get[tabs.TTabs](c.ref).InputValue(activeIdx.nago()))}
}

// InputValue binds the index of the current step to the state.
//
//gostrict:host
func (c TStepper) InputValue(state IndexState) TStepper {
	return TStepper{ref: wrap(get[stepper.TStepper](c.ref).InputValue(state.nago()))}
}

// InputValue binds the picked color to the state.
//
//gostrict:host
func (c TColorpickerPicker) InputValue(state ColorState) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).InputValue(state.nago()))}
}

// InputValue binds the picked color to the state.
//
//gostrict:host
func (c TColorpickerPickerField) InputValue(state ColorState) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).InputValue(state.nago()))}
}

// TPager shows the numbers of pages, like pager.TPager of nago.
type TPager struct{ ref js.Ref }

func (TPager) isView() {}

// Pager returns the numbers of pages, whose current index the state holds.
//
//gostrict:host
func Pager(pageIdx IndexState) TPager {
	return TPager{ref: wrap(pager.Pager(pageIdx.nago()))}
}

// Count sets the number of pages.
//
//gostrict:host
func (c TPager) Count(count int32) TPager {
	return TPager{ref: wrap(get[pager.TPager](c.ref).Count(int(count)))}
}

// Frame sets the size.
//
//gostrict:host
func (c TPager) Frame(frame Frame) TPager {
	return TPager{ref: wrap(get[pager.TPager](c.ref).Frame(frame.nago()))}
}

// Visible shows or hides the pager.
//
//gostrict:host
func (c TPager) Visible(v bool) TPager {
	return TPager{ref: wrap(get[pager.TPager](c.ref).Visible(v))}
}

// Switcher returns the pages, between which the user switches, and whose id the state holds, like
// switcher.Switcher of nago, which takes a slice of the pages as its first parameter.
//
//gostrict:host
func Switcher(state StringState, pages ...TSwitcherPage) TSwitcher {
	return TSwitcher{ref: wrap(switcher.Switcher(switcherPagesOf(pages), state.nago()))}
}

// HSwitcher returns pages, which are laid out horizontally, see Switcher.
//
//gostrict:host
func HSwitcher(state StringState, pages ...TSwitcherPage) TSwitcher {
	return TSwitcher{ref: wrap(switcher.HSwitcher(switcherPagesOf(pages), state.nago()))}
}

// VSwitcher returns pages, which are laid out vertically, see Switcher.
//
//gostrict:host
func VSwitcher(state StringState, pages ...TSwitcherPage) TSwitcher {
	return TSwitcher{ref: wrap(switcher.VSwitcher(switcherPagesOf(pages), state.nago()))}
}

// Chart is the common part of a chart, like chart.Chart of nago without its labels and colors: the labels of the axis
// follow the chart in the constructors, and the theme colors the series.
type Chart struct {
	Frame         Frame
	Downloadable  bool
	NoDataMessage string
	XAxisTitle    string
	YAxisTitle    string
	LabelRounding ChartRoundingType
}

// TChartSeries is a series of data points of a chart, like chart.Series of nago.
type TChartSeries struct{ ref js.Ref }

// ChartSeries returns a series with the label, the type and the points.
//
//gostrict:host
func ChartSeries(label string, typ ChartSeriesType, points ...ChartDataPoint) TChartSeries {
	return TChartSeries{ref: wrap(chartSeriesOf(label, typ, points))}
}

// BarChart returns a bar chart with the labels of its axis, like barchart.BarChart of nago.
//
//gostrict:host
func BarChart(c Chart, labels ...string) TBarChart {
	return TBarChart{ref: wrap(barchart.BarChart(c.nago(labels)))}
}

// Series sets the series.
//
//gostrict:host
func (c TBarChart) Series(series ...TChartSeries) TBarChart {
	return TBarChart{ref: wrap(get[barchart.TBarChart](c.ref).Series(chartSeriesList(series)))}
}

// LineChart returns a line chart with the labels of its axis, like linechart.LineChart of nago.
//
//gostrict:host
func LineChart(c Chart, labels ...string) TLineChart {
	return TLineChart{ref: wrap(linechart.LineChart(c.nago(labels)))}
}

// Series sets the series.
//
//gostrict:host
func (c TLineChart) Series(series ...TChartSeries) TLineChart {
	return TLineChart{ref: wrap(get[linechart.TLineChart](c.ref).Series(chartSeriesList(series)))}
}

// PieChart returns a pie chart with the labels of its slices, like piechart.PieChart of nago.
//
//gostrict:host
func PieChart(c Chart, labels ...string) TPieChart {
	return TPieChart{ref: wrap(piechart.PieChart(c.nago(labels)))}
}

// Series sets the series.
//
//gostrict:host
func (c TPieChart) Series(series ...TChartSeries) TPieChart {
	return TPieChart{ref: wrap(get[piechart.TPieChart](c.ref).Series(chartSeriesList(series)))}
}
