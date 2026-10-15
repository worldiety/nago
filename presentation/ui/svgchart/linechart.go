// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package svgchart

import (
	"math"
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/chart"
)

// Curve is the style of the line between two data points.
type Curve int

const (
	// CurveStraight connects the data points with straight lines.
	CurveStraight Curve = iota
	// CurveSmooth connects the data points with a monotone cubic spline, which never overshoots the values.
	CurveSmooth
	// CurveStepline holds each value until the next data point.
	CurveStepline
)

// Markers configures the dots on the data points of a line chart. A size of zero hides them.
type Markers struct {
	Size        int      // marker radius in px
	BorderColor ui.Color // marker border color
}

// TLineChart is a composite component (Line Chart).
// It shows how values develop over the categories and renders them as a static SVG image. A series of type
// [chart.ChartSeriesTypeArea] fills the area below its line, a series of type [chart.ChartSeriesTypeColumn]
// is drawn as bars behind the lines.
type TLineChart struct {
	chart          chart.Chart
	series         []chart.Series
	markers        Markers
	curve          Curve
	showDataLabels bool
}

// LineChart returns a TLineChart initialized with the given chart.
func LineChart(chart chart.Chart) TLineChart {
	return TLineChart{
		chart: chart,
	}
}

// Chart sets the chart config.
func (c TLineChart) Chart(chart chart.Chart) TLineChart {
	c.chart = chart
	return c
}

// Series sets the data series.
func (c TLineChart) Series(series []chart.Series) TLineChart {
	c.series = series
	return c
}

// Markers sets the markers config.
func (c TLineChart) Markers(markers Markers) TLineChart {
	c.markers = markers
	return c
}

// Curve sets the curve style.
func (c TLineChart) Curve(curve Curve) TLineChart {
	c.curve = curve
	return c
}

// ShowDataLabels shows the value above each data point.
func (c TLineChart) ShowDataLabels(show bool) TLineChart {
	c.showDataLabels = show
	return c
}

// SVG renders the chart with the default colors for a light theme.
func (c TLineChart) SVG() core.SVG {
	return c.svg(false)
}

type point struct {
	x, y  float64
	value float64
	cat   string
}

func (c TLineChart) svg(dark bool) core.SVG {
	width, height := frameSize(c.chart.Frame)
	if !hasData(c.series) {
		return noData(c.chart, width, height)
	}

	cats := categories(c.chart, c.series)
	lo, hi := minMax(c.series, cats, false)

	var legend []legendItem
	var columns []columnSeries
	for i, s := range c.series {
		color := seriesColor(c.chart, dark, i)
		legend = append(legend, legendItem{label: s.Label, color: color})
		if s.Type == chart.ChartSeriesTypeColumn {
			columns = append(columns, columnSeries{series: s, color: color, index: i})
		}
	}

	var w writer
	w.open(width, height)
	layout := newCartesian(c.chart, width, height, cats, lo, hi, false, legend)
	layout.drawAxes(&w)
	layout.drawBars(&w, columns, false, false)

	for i, s := range c.series {
		if s.Type == chart.ChartSeriesTypeColumn {
			continue
		}

		color := seriesColor(c.chart, dark, i)
		points := c.points(layout, s)
		if len(points) == 0 {
			continue
		}

		d := c.path(points)
		if s.Type == chart.ChartSeriesTypeArea && len(points) > 1 {
			base := layout.valuePos(0)
			area := d + "L" + num(points[len(points)-1].x) + " " + num(base) + "L" + num(points[0].x) + " " + num(base) + "Z"
			w.printf(`<path d="%s" fill="%s" fill-opacity="0.15"/>`, area, color)
		}

		if len(points) > 1 {
			w.printf(`<path d="%s" fill="none" stroke="%s" stroke-width="%s" stroke-linejoin="round" stroke-linecap="round"/>`, d, color, num(lineWidth))
		}

		c.drawPoints(&w, s.Label, points, color)
	}

	w.close()

	return w.bytes()
}

// points maps the data points onto the plot area in the order of the categories.
func (c TLineChart) points(layout *cartesian, s chart.Series) []point {
	byCat := make([]*point, len(layout.cats))
	for i, dp := range s.DataPoints {
		idx := categoryIndex(layout.cats, dp, i)
		if idx >= len(layout.cats) {
			continue
		}

		start, size := layout.band(idx)
		byCat[idx] = &point{x: start + size/2, y: layout.valuePos(dp.Y), value: dp.Y, cat: layout.cats[idx]}
	}

	var points []point
	for _, p := range byCat {
		if p != nil {
			points = append(points, *p)
		}
	}

	return points
}

// path returns the path data of the line through the points according to the curve.
func (c TLineChart) path(points []point) string {
	var sb strings.Builder
	sb.WriteString("M" + num(points[0].x) + " " + num(points[0].y))

	switch c.curve {
	case CurveStepline:
		for i := 1; i < len(points); i++ {
			sb.WriteString("H" + num(points[i].x) + "V" + num(points[i].y))
		}
	case CurveSmooth:
		tangents := monotoneTangents(points)
		for i := 1; i < len(points); i++ {
			p0, p1 := points[i-1], points[i]
			dx := (p1.x - p0.x) / 3
			sb.WriteString("C" + num(p0.x+dx) + " " + num(p0.y+tangents[i-1]*dx) + " " +
				num(p1.x-dx) + " " + num(p1.y-tangents[i]*dx) + " " +
				num(p1.x) + " " + num(p1.y))
		}
	default:
		for i := 1; i < len(points); i++ {
			sb.WriteString("L" + num(points[i].x) + " " + num(points[i].y))
		}
	}

	return sb.String()
}

// monotoneTangents returns the tangent slope at each point following Fritsch and Carlson, so that the
// interpolated curve does not overshoot the data.
func monotoneTangents(points []point) []float64 {
	n := len(points)
	tangents := make([]float64, n)
	if n < 2 {
		return tangents
	}

	slopes := make([]float64, n-1)
	for i := 0; i < n-1; i++ {
		dx := points[i+1].x - points[i].x
		if dx == 0 {
			slopes[i] = 0
			continue
		}

		slopes[i] = (points[i+1].y - points[i].y) / dx
	}

	tangents[0] = slopes[0]
	tangents[n-1] = slopes[n-2]
	for i := 1; i < n-1; i++ {
		if slopes[i-1]*slopes[i] <= 0 {
			tangents[i] = 0
		} else {
			tangents[i] = (slopes[i-1] + slopes[i]) / 2
		}
	}

	for i := 0; i < n-1; i++ {
		if slopes[i] == 0 {
			tangents[i] = 0
			tangents[i+1] = 0
			continue
		}

		a := tangents[i] / slopes[i]
		b := tangents[i+1] / slopes[i]
		s := a*a + b*b
		if s > 9 {
			t := 3 / math.Sqrt(s)
			tangents[i] = t * a * slopes[i]
			tangents[i+1] = t * b * slopes[i]
		}
	}

	return tangents
}

// drawPoints draws the markers and the data labels of the series.
func (c TLineChart) drawPoints(w *writer, label string, points []point, color string) {
	r := float64(c.markers.Size)
	if r <= 0 && !c.showDataLabels && len(points) > 1 {
		return
	}

	if r <= 0 {
		// a single point is invisible without a marker and labels need a hover target
		r = 3
	}

	stroke := ""
	if c.markers.BorderColor != "" {
		stroke = ` stroke="` + cssColor(c.markers.BorderColor) + `" stroke-width="2"`
	}

	for _, p := range points {
		value := formatValue(p.value, c.chart.LabelRounding)
		if c.markers.Size > 0 || len(points) == 1 {
			w.printf(`<circle cx="%s" cy="%s" r="%s" fill="%s"%s>`, num(p.x), num(p.y), num(r), color, stroke)
		} else {
			// invisible hover target
			w.printf(`<circle cx="%s" cy="%s" r="%s" fill="transparent">`, num(p.x), num(p.y), num(r*2))
		}

		w.title(tooltip(label, p.cat, value))
		w.sb.WriteString(`</circle>`)

		if c.showDataLabels {
			w.text(p.x, p.y-r-4, value, "middle", ` font-size="`+num(smallFontSize)+`" opacity="0.8"`)
		}
	}
}

// Render embeds the SVG as an image using the frame of the chart.
func (c TLineChart) Render(ctx core.RenderContext) core.RenderNode {
	return render(ctx, c.chart, c.svg(prefersDark(ctx)))
}
