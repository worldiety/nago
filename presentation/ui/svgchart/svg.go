// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package svgchart

import (
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/chart"
)

const (
	fontSize      = 12.0
	smallFontSize = 10.0
	// charWidth is the average glyph width of a sans-serif font relative to its size. It is only an estimation
	// used to reserve space for labels, as the server cannot measure the text.
	charWidth     = 0.58
	padding       = 16.0
	legendRowH    = 20.0
	legendSwatch  = 10.0
	defaultWidth  = 560.0
	defaultHeight = 320.0
	// maxBarThickness caps the bars, so that a chart with a few categories does not become a block of color.
	maxBarThickness = 32.0
	barRadius       = 4.0
	lineWidth       = 2.0
)

// lightPalette and darkPalette are the default series colors, if the chart declares no colors. They are
// ordered, so that neighbouring series stay distinguishable for color vision deficiencies. The dark palette
// contains the same hues stepped for a dark surface.
var (
	lightPalette = []ui.Color{"#2a78d6", "#eb6834", "#1baf7a", "#eda100", "#e87ba4", "#008300", "#6250d6", "#e34948"}
	darkPalette  = []ui.Color{"#3987e5", "#d95926", "#199e70", "#c98500", "#d55181", "#008300", "#9085e9", "#e66767"}
)

// writer collects the SVG markup.
type writer struct {
	sb strings.Builder
}

func (w *writer) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(&w.sb, format, args...)
}

// open writes the root element. The svg has no intrinsic size, so that the frame of the image decides about the
// rendered size and the viewBox scales the drawing.
func (w *writer) open(width, height float64) {
	w.printf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %s %s" preserveAspectRatio="xMidYMid meet" font-size="%s" fill="currentColor">`, num(width), num(height), num(fontSize))
}

func (w *writer) close() {
	w.sb.WriteString(`</svg>`)
}

// text writes a text element. The anchor is one of start, middle or end.
func (w *writer) text(x, y float64, s string, anchor string, extra string) {
	w.printf(`<text x="%s" y="%s" text-anchor="%s"%s>%s</text>`, num(x), num(y), anchor, extra, esc(s))
}

func (w *writer) title(s string) {
	w.printf(`<title>%s</title>`, esc(s))
}

func (w *writer) line(x1, y1, x2, y2 float64, extra string) {
	w.printf(`<line x1="%s" y1="%s" x2="%s" y2="%s"%s/>`, num(x1), num(y1), num(x2), num(y2), extra)
}

func (w *writer) bytes() core.SVG {
	return core.SVG(w.sb.String())
}

// num formats a coordinate with at most two decimals.
func num(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}

	s := strconv.FormatFloat(v, 'f', 2, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}

	if s == "-0" {
		return "0"
	}

	return s
}

func esc(s string) string {
	return html.EscapeString(s)
}

// textWidth estimates the rendered width of the text.
func textWidth(s string, size float64) float64 {
	return float64(utf8.RuneCountInString(s)) * size * charWidth
}

// cssColor converts a nago color into a CSS color. Named colors become references to the theme variables,
// which the frontend defines on the document, see the colorValue function of the web frontend.
func cssColor(c ui.Color) string {
	s := string(c)
	if s == "" {
		return "currentColor"
	}

	name, alpha, hasAlpha := strings.Cut(s, "/")
	value := name
	if !strings.HasPrefix(name, "#") && !strings.HasPrefix(name, "rgb") {
		value = "var(--" + name + ")"
	}

	if !hasAlpha {
		return value
	}

	a, err := strconv.Atoi(alpha)
	if err != nil {
		return value
	}

	return fmt.Sprintf("color-mix(in srgb, %s %s%%, transparent)", value, num(float64(a)/255*100))
}

// formatValue formats a data value for a label according to the rounding of the chart.
func formatValue(v float64, rounding chart.RoundingType) string {
	switch rounding {
	case chart.RoundingRound:
		v = math.Round(v)
	case chart.RoundingFloor:
		v = math.Floor(v)
	case chart.RoundingCeiling:
		v = math.Ceil(v)
	}

	return num(v)
}

// frameSize derives the drawing size from the frame. Only absolute lengths in dp or rem are known, everything
// else falls back to the default size and is scaled by the browser.
func frameSize(f ui.Frame) (float64, float64) {
	w := absoluteLength(f.Width)
	if w <= 0 {
		w = absoluteLength(f.MaxWidth)
	}

	h := absoluteLength(f.Height)
	if h <= 0 {
		h = absoluteLength(f.MaxHeight)
	}

	if w <= 0 && h <= 0 {
		return defaultWidth, defaultHeight
	}

	if w <= 0 {
		w = h * defaultWidth / defaultHeight
	}

	if h <= 0 {
		h = w * defaultHeight / defaultWidth
	}

	return w, h
}

func absoluteLength(l ui.Length) float64 {
	s := string(l)
	for _, unit := range []string{"dp", "rem", "px"} {
		if !strings.HasSuffix(s, unit) {
			continue
		}

		v, err := strconv.ParseFloat(strings.TrimSuffix(s, unit), 64)
		if err != nil {
			return 0
		}

		if unit == "rem" {
			v *= 16
		}

		return v
	}

	return 0
}

// seriesColor returns the color of the series with the given index.
func seriesColor(cfg chart.Chart, dark bool, idx int) string {
	colors := cfg.Colors
	if len(colors) == 0 {
		colors = lightPalette
		if dark {
			colors = darkPalette
		}
	}

	return cssColor(colors[idx%len(colors)])
}

// categories returns the labels of the category axis. The labels of the chart win, otherwise the X values of
// all series are collected in the order of their first appearance.
func categories(cfg chart.Chart, series []chart.Series) []string {
	if len(cfg.Labels) > 0 {
		return cfg.Labels
	}

	var cats []string
	seen := map[string]bool{}
	for _, s := range series {
		for _, dp := range s.DataPoints {
			if !seen[dp.X] {
				seen[dp.X] = true
				cats = append(cats, dp.X)
			}
		}
	}

	return cats
}

// categoryIndex returns the category of the data point. A data point whose X value is not a category is placed
// by its position within the series.
func categoryIndex(cats []string, dp chart.DataPoint, pos int) int {
	for i, c := range cats {
		if c == dp.X {
			return i
		}
	}

	return pos
}

func hasData(series []chart.Series) bool {
	for _, s := range series {
		if len(s.DataPoints) > 0 {
			return true
		}
	}

	return false
}

// noData renders an empty chart with the message of the chart, if any.
func noData(cfg chart.Chart, width, height float64) core.SVG {
	var w writer
	w.open(width, height)
	if cfg.NoDataMessage != "" {
		w.text(width/2, height/2, cfg.NoDataMessage, "middle", ` dominant-baseline="middle" opacity="0.7"`)
	}

	w.close()
	return w.bytes()
}

// render embeds the svg as image using the frame of the chart. Without a frame the image takes the full width
// and keeps the aspect ratio of the drawing.
func render(ctx core.RenderContext, cfg chart.Chart, svg core.SVG) core.RenderNode {
	frame := cfg.Frame
	if frame.IsZero() {
		frame.Width = ui.Full
	}

	return ui.Image().Embed(svg).Frame(frame).Render(ctx)
}

func prefersDark(ctx core.RenderContext) bool {
	if ctx == nil || ctx.Window() == nil {
		return false
	}

	return ctx.Window().Info().PrefersDark()
}
