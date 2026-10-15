// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package svgchart

import (
	"log/slog"
	"math"
	"strconv"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/chart"
)

// Marker highlights a target value on a single bar, e.g. a goal. It is drawn as a line across the bar at
// the given value. Width and Height are the thickness of the line, depending on the orientation of the chart,
// Round draws a dot instead of a line and Dashed a dashed line. The label is shown as tooltip.
type Marker struct {
	Label          string
	SeriesIndex    int
	DataPointIndex int
	Value          string
	Width          int
	Height         int
	Color          ui.Color
	Round          bool
	Dashed         bool
}

// TBarChart is a composite component (Bar Chart).
// It compares values across categories with rectangular bars, grouped or stacked, vertical or horizontal,
// and renders them as a static SVG image.
type TBarChart struct {
	chart          chart.Chart
	series         []chart.Series
	markers        []Marker
	horizontal     bool
	stacked        bool
	showDataLabels bool
}

// BarChart creates a new bar chart with default (vertical, non-stacked) configuration and data labels.
func BarChart(chart chart.Chart) TBarChart {
	return TBarChart{
		chart:          chart,
		showDataLabels: true,
	}
}

// Chart sets the underlying chart configuration for the bar chart.
func (c TBarChart) Chart(chart chart.Chart) TBarChart {
	c.chart = chart
	return c
}

// Series defines the data series to be displayed in the bar chart.
func (c TBarChart) Series(series []chart.Series) TBarChart {
	c.series = series
	return c
}

// Markers adds markers to the bar chart to highlight values. Markers are not supported for stacked charts.
func (c TBarChart) Markers(markers []Marker) TBarChart {
	c.markers = markers
	return c
}

// Horizontal sets whether the bar chart is rendered horizontally.
func (c TBarChart) Horizontal(horizontal bool) TBarChart {
	c.horizontal = horizontal
	return c
}

// Stacked sets whether multiple series are stacked instead of grouped.
func (c TBarChart) Stacked(stacked bool) TBarChart {
	c.stacked = stacked
	return c
}

// ShowDataLabels shows or hides the values at the end of the bars. Stacked bars show the total of the stack.
func (c TBarChart) ShowDataLabels(show bool) TBarChart {
	c.showDataLabels = show
	return c
}

// SVG renders the chart with the default colors for a light theme.
func (c TBarChart) SVG() core.SVG {
	return c.svg(false)
}

func (c TBarChart) svg(dark bool) core.SVG {
	width, height := frameSize(c.chart.Frame)
	if !hasData(c.series) {
		return noData(c.chart, width, height)
	}

	cats := categories(c.chart, c.series)
	lo, hi := minMax(c.series, cats, c.stacked)

	var legend []legendItem
	var columns []columnSeries
	for i, s := range c.series {
		color := seriesColor(c.chart, dark, i)
		legend = append(legend, legendItem{label: s.Label, color: color})
		columns = append(columns, columnSeries{series: s, color: color, index: i})
	}

	var w writer
	w.open(width, height)
	layout := newCartesian(c.chart, width, height, cats, lo, hi, c.horizontal, legend)
	layout.drawAxes(&w)
	bars := layout.drawBars(&w, columns, c.stacked, c.showDataLabels)

	if c.stacked {
		if len(c.markers) > 0 {
			slog.Warn("svgchart: markers are not supported for stacked bar charts")
		}
	} else {
		c.drawMarkers(&w, layout, bars)
	}

	w.close()

	return w.bytes()
}

func (c TBarChart) drawMarkers(w *writer, layout *cartesian, bars []bar) {
	for _, m := range c.markers {
		value, err := strconv.ParseFloat(m.Value, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			slog.Warn("svgchart: marker value is not a number", "value", m.Value)
			continue
		}

		var target *bar
		for i := range bars {
			if bars[i].seriesIndex == m.SeriesIndex && bars[i].dataPointIndex == m.DataPointIndex {
				target = &bars[i]
				break
			}
		}

		if target == nil {
			continue
		}

		thickness := float64(m.Height)
		if c.horizontal {
			thickness = float64(m.Width)
		}

		if thickness <= 0 {
			thickness = 2
		}

		pos := layout.valuePos(math.Min(math.Max(value, layout.scale.min), layout.scale.max))
		color := cssColor(m.Color)
		center := target.start + target.thickness/2

		if m.Round {
			w.printf(`<circle cx="%s" cy="%s" r="%s" fill="%s">`, num(markerX(c.horizontal, pos, center)), num(markerY(c.horizontal, pos, center)), num(thickness/2), color)
			w.title(markerTitle(m))
			w.sb.WriteString(`</circle>`)
			continue
		}

		extra := ` stroke="` + color + `" stroke-width="` + num(thickness) + `"`
		if m.Dashed {
			extra += ` stroke-dasharray="3 3"`
		}

		w.printf(`<g%s>`, extra)
		if c.horizontal {
			w.line(pos, target.start, pos, target.start+target.thickness, "")
		} else {
			w.line(target.start, pos, target.start+target.thickness, pos, "")
		}

		w.title(markerTitle(m))
		w.sb.WriteString(`</g>`)
	}
}

func markerX(horizontal bool, pos, center float64) float64 {
	if horizontal {
		return pos
	}

	return center
}

func markerY(horizontal bool, pos, center float64) float64 {
	if horizontal {
		return center
	}

	return pos
}

func markerTitle(m Marker) string {
	if m.Label == "" {
		return m.Value
	}

	return m.Label + ": " + m.Value
}

// Render embeds the SVG as an image using the frame of the chart.
func (c TBarChart) Render(ctx core.RenderContext) core.RenderNode {
	return render(ctx, c.chart, c.svg(prefersDark(ctx)))
}
