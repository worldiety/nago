// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package svgchart

import (
	"math"
	"strconv"

	"go.wdy.de/nago/presentation/ui/chart"
)

// scale maps values onto an axis. The axis always contains zero and its ticks are rounded to nice numbers.
type scale struct {
	min, max float64
	ticks    []float64
	decimals int
}

// niceScale returns a scale which covers lo..hi and zero.
func niceScale(lo, hi float64) scale {
	lo = math.Min(lo, 0)
	hi = math.Max(hi, 0)
	if hi-lo == 0 {
		hi = 1
	}

	step := niceNum((hi-lo)/4, true)
	var s scale
	if step < 1 {
		s.decimals = int(math.Ceil(-math.Log10(step)))
	}

	// multiples of the step are rounded, so that e.g. 3*0.1 does not end up as 0.30000000000000004
	s.min = s.round(math.Floor(lo/step) * step)
	s.max = s.round(math.Ceil(hi/step) * step)
	for i := 0; ; i++ {
		v := s.round(s.min + float64(i)*step)
		if v > s.max {
			break
		}

		s.ticks = append(s.ticks, v)
	}

	return s
}

func (s scale) round(v float64) float64 {
	f := math.Pow(10, float64(s.decimals))
	return math.Round(v*f) / f
}

// niceNum rounds x to 1, 2, 5 or 10 times a power of ten, see Heckbert's nice numbers algorithm.
func niceNum(x float64, round bool) float64 {
	exp := math.Floor(math.Log10(x))
	f := x / math.Pow(10, exp)

	var nf float64
	switch {
	case round && f < 1.5, !round && f <= 1:
		nf = 1
	case round && f < 3, !round && f <= 2:
		nf = 2
	case round && f < 7, !round && f <= 5:
		nf = 5
	default:
		nf = 10
	}

	return nf * math.Pow(10, exp)
}

// fraction returns the position of v within the scale as 0..1.
func (s scale) fraction(v float64) float64 {
	return (v - s.min) / (s.max - s.min)
}

func (s scale) tickLabel(v float64) string {
	if math.Abs(v) < 1e-9 {
		return "0"
	}

	return strconv.FormatFloat(v, 'f', s.decimals, 64)
}

// minMax returns the smallest and largest value over all series. Stacked charts sum the positive and
// the negative values of a category separately.
func minMax(series []chart.Series, cats []string, stacked bool) (float64, float64) {
	lo, hi := 0.0, 0.0
	if !stacked {
		for _, s := range series {
			for _, dp := range s.DataPoints {
				lo = math.Min(lo, dp.Y)
				hi = math.Max(hi, dp.Y)
			}
		}

		return lo, hi
	}

	pos := make([]float64, len(cats))
	neg := make([]float64, len(cats))
	for _, s := range series {
		for i, dp := range s.DataPoints {
			idx := categoryIndex(cats, dp, i)
			if idx >= len(cats) {
				continue
			}

			if dp.Y >= 0 {
				pos[idx] += dp.Y
			} else {
				neg[idx] += dp.Y
			}
		}
	}

	for i := range cats {
		lo = math.Min(lo, neg[i])
		hi = math.Max(hi, pos[i])
	}

	return lo, hi
}

type legendItem struct {
	label string
	color string
}

// legend arranges the items into rows which fit into the width and returns the rows.
func legendRows(items []legendItem, width float64) [][]legendItem {
	var rows [][]legendItem
	var row []legendItem
	x := 0.0
	for _, item := range items {
		w := legendItemWidth(item)
		if len(row) > 0 && x+w > width {
			rows = append(rows, row)
			row = nil
			x = 0
		}

		row = append(row, item)
		x += w
	}

	if len(row) > 0 {
		rows = append(rows, row)
	}

	return rows
}

func legendItemWidth(item legendItem) float64 {
	return legendSwatch + 6 + textWidth(item.label, fontSize) + 12
}

// drawLegend draws the rows centered, starting at the given y coordinate.
func drawLegend(w *writer, rows [][]legendItem, width, y float64) {
	for _, row := range rows {
		total := 0.0
		for _, item := range row {
			total += legendItemWidth(item)
		}

		x := (width - total) / 2
		for _, item := range row {
			w.printf(`<circle cx="%s" cy="%s" r="%s" fill="%s"/>`, num(x+legendSwatch/2), num(y), num(legendSwatch/2), item.color)
			w.text(x+legendSwatch+6, y, item.label, "start", ` dominant-baseline="middle"`)
			x += legendItemWidth(item)
		}

		y += legendRowH
	}
}

// cartesian is the layout of a chart with a category axis and a value axis.
type cartesian struct {
	cfg        chart.Chart
	width      float64
	height     float64
	cats       []string
	scale      scale
	horizontal bool
	legend     [][]legendItem

	// the plot area
	x, y, w, h float64
}

// newCartesian computes the plot area so that the tick labels, the category labels, the axis titles and the
// legend fit around it.
func newCartesian(cfg chart.Chart, width, height float64, cats []string, lo, hi float64, horizontal bool, legend []legendItem) *cartesian {
	c := &cartesian{
		cfg:        cfg,
		width:      width,
		height:     height,
		cats:       cats,
		scale:      niceScale(lo, hi),
		horizontal: horizontal,
	}

	if len(legend) > 1 {
		c.legend = legendRows(legend, width-2*padding)
	}

	tickWidth := 0.0
	for _, t := range c.scale.ticks {
		tickWidth = math.Max(tickWidth, textWidth(c.scale.tickLabel(t), fontSize))
	}

	catWidth := 0.0
	for _, cat := range cats {
		catWidth = math.Max(catWidth, textWidth(cat, fontSize))
	}

	left := padding + 8
	if horizontal {
		left += math.Min(catWidth, width/3)
	} else {
		left += tickWidth
	}

	if cfg.YAxisTitle != "" {
		left += fontSize + 8
	}

	bottom := padding + fontSize + 8
	if cfg.XAxisTitle != "" {
		bottom += fontSize + 8
	}

	bottom += float64(len(c.legend)) * legendRowH

	c.x = left
	c.y = padding
	c.w = math.Max(width-left-padding, 10)
	c.h = math.Max(height-c.y-bottom, 10)

	return c
}

// valuePos returns the pixel position of the value on the value axis.
func (c *cartesian) valuePos(v float64) float64 {
	f := c.scale.fraction(v)
	if c.horizontal {
		return c.x + f*c.w
	}

	return c.y + c.h - f*c.h
}

// band returns the start and the size of the category band.
func (c *cartesian) band(idx int) (float64, float64) {
	n := float64(max(len(c.cats), 1))
	if c.horizontal {
		size := c.h / n
		return c.y + float64(idx)*size, size
	}

	size := c.w / n
	return c.x + float64(idx)*size, size
}

// drawAxes draws the grid lines, the tick labels, the category labels and the axis titles.
func (c *cartesian) drawAxes(w *writer) {
	// grid and tick labels of the value axis
	for _, t := range c.scale.ticks {
		p := c.valuePos(t)
		label := c.scale.tickLabel(t)
		if c.horizontal {
			w.line(p, c.y, p, c.y+c.h, ` stroke="currentColor" stroke-opacity="0.15"`)
			w.text(p, c.y+c.h+fontSize+4, label, "middle", ` opacity="0.7"`)
		} else {
			w.line(c.x, p, c.x+c.w, p, ` stroke="currentColor" stroke-opacity="0.15"`)
			w.text(c.x-8, p, label, "end", ` dominant-baseline="middle" opacity="0.7"`)
		}
	}

	// base line at zero
	zero := c.valuePos(0)
	if c.horizontal {
		w.line(zero, c.y, zero, c.y+c.h, ` stroke="currentColor" stroke-opacity="0.4"`)
	} else {
		w.line(c.x, zero, c.x+c.w, zero, ` stroke="currentColor" stroke-opacity="0.4"`)
	}

	// category labels, skipping some if they would overlap
	every := 1
	if !c.horizontal && len(c.cats) > 0 {
		maxWidth := 0.0
		for _, cat := range c.cats {
			maxWidth = math.Max(maxWidth, textWidth(cat, fontSize)+8)
		}

		_, size := c.band(0)
		if size > 0 {
			every = max(1, int(math.Ceil(maxWidth/size)))
		}
	} else if c.horizontal && len(c.cats) > 0 {
		_, size := c.band(0)
		if size > 0 {
			every = max(1, int(math.Ceil((fontSize+4)/size)))
		}
	}

	for i, cat := range c.cats {
		if i%every != 0 {
			continue
		}

		start, size := c.band(i)
		if c.horizontal {
			w.text(c.x-8, start+size/2, truncate(cat, c.categoryChars()), "end", ` dominant-baseline="middle" opacity="0.7"`)
		} else {
			w.text(start+size/2, c.y+c.h+fontSize+4, cat, "middle", ` opacity="0.7"`)
		}
	}

	if c.cfg.XAxisTitle != "" {
		y := c.y + c.h + 2*fontSize + 12
		w.text(c.x+c.w/2, y, c.cfg.XAxisTitle, "middle", ` font-weight="600"`)
	}

	if c.cfg.YAxisTitle != "" {
		x := padding + fontSize/2
		y := c.y + c.h/2
		w.printf(`<text x="%s" y="%s" text-anchor="middle" dominant-baseline="middle" font-weight="600" transform="rotate(-90 %s %s)">%s</text>`, num(x), num(y), num(x), num(y), esc(c.cfg.YAxisTitle))
	}

	if len(c.legend) > 0 {
		drawLegend(w, c.legend, c.width, c.height-padding-float64(len(c.legend)-1)*legendRowH-legendRowH/2)
	}
}

// categoryChars returns the number of characters which fit left of the plot area of a horizontal chart.
func (c *cartesian) categoryChars() int {
	space := c.x - padding - 8
	if c.cfg.YAxisTitle != "" {
		space -= fontSize + 8
	}

	return int(math.Round(space / (fontSize * charWidth)))
}

// truncate shortens the string to n characters.
func truncate(s string, n int) string {
	if n < 2 {
		n = 2
	}

	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	return string(runes[:n-1]) + "…"
}

// barPath returns a path of a bar whose end away from the baseline is rounded. The bar spans from the
// position base to end on the value axis and has the given thickness, starting at start on the category axis.
func (c *cartesian) barPath(start, thickness, base, end float64, rounded bool) string {
	r := 0.0
	if rounded {
		r = math.Min(barRadius, math.Min(thickness/2, math.Abs(end-base)))
	}

	if c.horizontal {
		// start is y, values are x
		y1, y2 := start, start+thickness
		if end >= base {
			// bar grows to the right, round the right corners
			return "M" + num(base) + " " + num(y1) +
				"H" + num(end-r) +
				"A" + num(r) + " " + num(r) + " 0 0 1 " + num(end) + " " + num(y1+r) +
				"V" + num(y2-r) +
				"A" + num(r) + " " + num(r) + " 0 0 1 " + num(end-r) + " " + num(y2) +
				"H" + num(base) + "Z"
		}

		return "M" + num(base) + " " + num(y1) +
			"H" + num(end+r) +
			"A" + num(r) + " " + num(r) + " 0 0 0 " + num(end) + " " + num(y1+r) +
			"V" + num(y2-r) +
			"A" + num(r) + " " + num(r) + " 0 0 0 " + num(end+r) + " " + num(y2) +
			"H" + num(base) + "Z"
	}

	x1, x2 := start, start+thickness
	if end <= base {
		// bar grows upwards, round the top corners
		return "M" + num(x1) + " " + num(base) +
			"V" + num(end+r) +
			"A" + num(r) + " " + num(r) + " 0 0 1 " + num(x1+r) + " " + num(end) +
			"H" + num(x2-r) +
			"A" + num(r) + " " + num(r) + " 0 0 1 " + num(x2) + " " + num(end+r) +
			"V" + num(base) + "Z"
	}

	return "M" + num(x1) + " " + num(base) +
		"V" + num(end-r) +
		"A" + num(r) + " " + num(r) + " 0 0 0 " + num(x1+r) + " " + num(end) +
		"H" + num(x2-r) +
		"A" + num(r) + " " + num(r) + " 0 0 0 " + num(x2) + " " + num(end-r) +
		"V" + num(base) + "Z"
}

// columnSeries is a series which is drawn as bars.
type columnSeries struct {
	series chart.Series
	color  string
	index  int
}

// bar describes a drawn bar, so that markers and labels can be placed on it.
type bar struct {
	seriesIndex    int
	dataPointIndex int
	category       int
	start          float64 // position on the category axis
	thickness      float64
	base, end      float64 // positions on the value axis
	value          float64
}

// drawBars draws the columns grouped or stacked and returns the bars.
func (c *cartesian) drawBars(w *writer, columns []columnSeries, stacked bool, showLabels bool) []bar {
	if len(columns) == 0 || len(c.cats) == 0 {
		return nil
	}

	var bars []bar
	_, bandSize := c.band(0)
	inner := bandSize * 0.7
	groups := len(columns)
	if stacked {
		groups = 1
	}

	gap := 2.0
	thickness := math.Min((inner-gap*float64(groups-1))/float64(groups), maxBarThickness)
	groupWidth := thickness*float64(groups) + gap*float64(groups-1)

	posTop := make([]float64, len(c.cats))
	negTop := make([]float64, len(c.cats))
	lastPos := make([]int, len(c.cats)) // index of the bar which is currently on top of the stack
	lastNeg := make([]int, len(c.cats))
	for i := range lastPos {
		lastPos[i], lastNeg[i] = -1, -1
	}

	for k, col := range columns {
		for i, dp := range col.series.DataPoints {
			idx := categoryIndex(c.cats, dp, i)
			if idx >= len(c.cats) {
				continue
			}

			bandStart, _ := c.band(idx)
			start := bandStart + (bandSize-groupWidth)/2
			if !stacked {
				start += float64(k) * (thickness + gap)
			}

			var base, end float64
			if stacked {
				if dp.Y >= 0 {
					base = c.valuePos(posTop[idx])
					posTop[idx] += dp.Y
					end = c.valuePos(posTop[idx])
				} else {
					base = c.valuePos(negTop[idx])
					negTop[idx] += dp.Y
					end = c.valuePos(negTop[idx])
				}
			} else {
				base = c.valuePos(0)
				end = c.valuePos(dp.Y)
			}

			b := bar{
				seriesIndex:    col.index,
				dataPointIndex: i,
				category:       idx,
				start:          start,
				thickness:      thickness,
				base:           base,
				end:            end,
				value:          dp.Y,
			}
			bars = append(bars, b)

			if stacked {
				if dp.Y >= 0 {
					lastPos[idx] = len(bars) - 1
				} else {
					lastNeg[idx] = len(bars) - 1
				}
			}
		}
	}

	// the outermost segment of a stack is rounded, all other segments keep a small gap to their neighbour
	rounded := make([]bool, len(bars))
	if stacked {
		for i := range c.cats {
			if lastPos[i] >= 0 {
				rounded[lastPos[i]] = true
			}
			if lastNeg[i] >= 0 {
				rounded[lastNeg[i]] = true
			}
		}
	} else {
		for i := range rounded {
			rounded[i] = true
		}
	}

	colorOf := map[int]string{}
	labelOf := map[int]string{}
	for _, col := range columns {
		colorOf[col.index] = col.color
		labelOf[col.index] = col.series.Label
	}

	for i, b := range bars {
		base, end := b.base, b.end
		if stacked && !rounded[i] && base != end {
			// leave a gap in the surface color between the segments
			if end < base {
				end += gap
			} else {
				end -= gap
			}
		}

		if math.Abs(end-base) < 0.01 {
			continue
		}

		w.printf(`<path d="%s" fill="%s">`, c.barPath(b.start, b.thickness, base, end, rounded[i]), colorOf[b.seriesIndex])
		w.title(tooltip(labelOf[b.seriesIndex], c.cats[b.category], formatValue(b.value, c.cfg.LabelRounding)))
		w.sb.WriteString(`</path>`)
	}

	if showLabels {
		c.drawBarLabels(w, bars, rounded, stacked, posTop, negTop)
	}

	return bars
}

// drawBarLabels writes the values at the ends of the bars. Stacked bars get the total of the stack. Labels
// which would overlap their neighbours are left out altogether, as the tooltips still carry the values.
func (c *cartesian) drawBarLabels(w *writer, bars []bar, rounded []bool, stacked bool, posTop, negTop []float64) {
	labels := make([]string, len(bars))
	for i, b := range bars {
		if !rounded[i] {
			continue
		}

		value := b.value
		if stacked {
			if b.value >= 0 {
				value = posTop[b.category]
			} else {
				value = negTop[b.category]
			}
		}

		labels[i] = formatValue(value, c.cfg.LabelRounding)
		if !c.horizontal && textWidth(labels[i], smallFontSize) > b.thickness+4 {
			return
		}
	}

	for i, b := range bars {
		if labels[i] == "" {
			continue
		}

		// negative bars get their label next to the baseline, where nothing else is drawn in their band
		extra := ` font-size="` + num(smallFontSize) + `" opacity="0.8"`
		if c.horizontal {
			x := b.end + 4
			if b.end < b.base {
				x = b.base + 4
			}

			w.text(x, b.start+b.thickness/2, labels[i], "start", extra+` dominant-baseline="middle"`)
			continue
		}

		y := b.end - 4
		if b.end > b.base {
			y = b.base - 4
		}

		w.text(b.start+b.thickness/2, y, labels[i], "middle", extra)
	}
}

func tooltip(series, category, value string) string {
	if series == "" {
		return category + ": " + value
	}

	return series + " · " + category + ": " + value
}
