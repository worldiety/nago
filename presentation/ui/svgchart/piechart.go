// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package svgchart

import (
	"math"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/chart"
)

// TPieChart is a composite component (Pie Chart).
// It shows the share of each data point of the first series in the whole and renders it as a static SVG image.
type TPieChart struct {
	chart              chart.Chart
	series             []chart.Series
	showAsDonut        bool
	showDataLabels     bool
	showAbsoluteValues bool
}

// PieChart creates a pie chart with the given chart configuration. Add the data with Series.
func PieChart(chart chart.Chart) TPieChart {
	return TPieChart{
		chart:          chart,
		showDataLabels: true,
	}
}

// Chart sets the chart configuration.
func (c TPieChart) Chart(chart chart.Chart) TPieChart {
	c.chart = chart
	return c
}

// Series sets the data series. Only the first series is rendered.
func (c TPieChart) Series(series []chart.Series) TPieChart {
	c.series = series
	return c
}

// ShowAsDonut renders a donut instead of a pie.
func (c TPieChart) ShowAsDonut(showAsDonut bool) TPieChart {
	c.showAsDonut = showAsDonut
	return c
}

// ShowDataLabels shows or hides the labels on the slices.
func (c TPieChart) ShowDataLabels(showDataLabels bool) TPieChart {
	c.showDataLabels = showDataLabels
	return c
}

// ShowAbsoluteValues shows the values instead of percentages.
func (c TPieChart) ShowAbsoluteValues(showAbsoluteValues bool) TPieChart {
	c.showAbsoluteValues = showAbsoluteValues
	return c
}

// SVG renders the chart with the default colors for a light theme.
func (c TPieChart) SVG() core.SVG {
	return c.svg(false)
}

type slice struct {
	label   string
	value   float64
	color   string
	percent float64
}

func (c TPieChart) svg(dark bool) core.SVG {
	width, height := frameSize(c.chart.Frame)
	slices := c.slices(dark)
	if len(slices) == 0 {
		return noData(c.chart, width, height)
	}

	legend := make([]legendItem, 0, len(slices))
	for _, s := range slices {
		legend = append(legend, legendItem{label: s.label, color: s.color})
	}

	rows := legendRows(legend, width-2*padding)
	legendHeight := float64(len(rows)) * legendRowH

	var w writer
	w.open(width, height)

	cx := width / 2
	plotHeight := height - 2*padding - legendHeight
	cy := padding + plotHeight/2
	r := math.Max(math.Min(width-2*padding, plotHeight)/2, 1)
	inner := 0.0
	if c.showAsDonut {
		inner = r * 0.6
	}

	start := -math.Pi / 2
	for _, s := range slices {
		angle := s.percent / 100 * 2 * math.Pi
		w.printf(`<path d="%s" fill="%s">`, arcPath(cx, cy, r, inner, start, start+angle), s.color)
		w.title(s.label + ": " + formatValue(s.value, c.chart.LabelRounding) + " (" + c.percentLabel(s) + ")")
		w.sb.WriteString(`</path>`)

		if c.showDataLabels && s.percent >= 4 {
			mid := start + angle/2
			lr := r * 0.62
			if c.showAsDonut {
				lr = (r + inner) / 2
			}

			label := c.percentLabel(s)
			if c.showAbsoluteValues {
				label = formatValue(s.value, c.chart.LabelRounding)
			}

			// white text with a dark outline is readable on any slice color
			w.text(cx+math.Cos(mid)*lr, cy+math.Sin(mid)*lr, label, "middle", ` dominant-baseline="middle" fill="#ffffff" stroke="rgba(0,0,0,0.4)" stroke-width="2" paint-order="stroke" font-weight="600"`)
		}

		start += angle
	}

	drawLegend(&w, rows, width, height-padding-legendHeight+legendRowH/2)
	w.close()

	return w.bytes()
}

func (c TPieChart) percentLabel(s slice) string {
	rounding := c.chart.LabelRounding
	if rounding == chart.RoundingNone {
		rounding = chart.RoundingRound
	}

	return formatValue(s.percent, rounding) + "%"
}

// slices returns the positive data points of the first series. Labels of the chart override the X values.
func (c TPieChart) slices(dark bool) []slice {
	if len(c.series) == 0 {
		return nil
	}

	total := 0.0
	for _, dp := range c.series[0].DataPoints {
		if dp.Y > 0 {
			total += dp.Y
		}
	}

	if total <= 0 {
		return nil
	}

	var slices []slice
	for i, dp := range c.series[0].DataPoints {
		if dp.Y <= 0 {
			continue
		}

		label := dp.X
		if i < len(c.chart.Labels) {
			label = c.chart.Labels[i]
		}

		slices = append(slices, slice{
			label:   label,
			value:   dp.Y,
			color:   seriesColor(c.chart, dark, i),
			percent: dp.Y / total * 100,
		})
	}

	return slices
}

// arcPath returns the path of a ring segment between the angles a0 and a1 in radians, clockwise from the
// center. An inner radius of zero yields a pie slice.
func arcPath(cx, cy, r, inner, a0, a1 float64) string {
	full := a1-a0 >= 2*math.Pi-1e-6
	if full {
		// a single arc cannot describe a full circle, so draw two halves
		a1 = a0 + 2*math.Pi
		mid := a0 + math.Pi
		outer := "M" + pt(cx, cy, r, a0) + arc(r, 0, 1, cx, cy, r, mid) + arc(r, 0, 1, cx, cy, r, a1)
		if inner <= 0 {
			return outer + "Z"
		}

		return outer + "M" + pt(cx, cy, inner, a1) + arc(inner, 0, 0, cx, cy, inner, mid) + arc(inner, 0, 0, cx, cy, inner, a0) + "Z"
	}

	large := 0
	if a1-a0 > math.Pi {
		large = 1
	}

	d := "M" + pt(cx, cy, r, a0) + arc(r, large, 1, cx, cy, r, a1)
	if inner <= 0 {
		return d + "L" + num(cx) + " " + num(cy) + "Z"
	}

	return d + "L" + pt(cx, cy, inner, a1) + arc(inner, large, 0, cx, cy, inner, a0) + "Z"
}

func pt(cx, cy, r, a float64) string {
	return num(cx+math.Cos(a)*r) + " " + num(cy+math.Sin(a)*r)
}

func arc(r float64, large, sweep int, cx, cy, radius, a float64) string {
	return "A" + num(r) + " " + num(r) + " 0 " + num(float64(large)) + " " + num(float64(sweep)) + " " + pt(cx, cy, radius, a)
}

// Render embeds the SVG as an image using the frame of the chart.
func (c TPieChart) Render(ctx core.RenderContext) core.RenderNode {
	return render(ctx, c.chart, c.svg(prefersDark(ctx)))
}
