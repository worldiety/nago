// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package svgchart

import (
	"encoding/xml"
	"strings"
	"testing"

	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/chart"
)

var testSeries = []chart.Series{
	{
		Label: "Online",
		DataPoints: []chart.DataPoint{
			{X: "Q1", Y: 120}, {X: "Q2", Y: 180}, {X: "Q3", Y: 150}, {X: "Q4", Y: 210},
		},
	},
	{
		Label: "Retail",
		DataPoints: []chart.DataPoint{
			{X: "Q1", Y: 90}, {X: "Q2", Y: 110}, {X: "Q3", Y: 130}, {X: "Q4", Y: 160.5},
		},
	},
}

func assertWellFormed(t *testing.T, svg string) {
	t.Helper()

	if !strings.HasPrefix(svg, `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Fatalf("not an svg: %s", svg)
	}

	dec := xml.NewDecoder(strings.NewReader(svg))
	for {
		_, err := dec.Token()
		if err != nil {
			if err.Error() == "EOF" {
				return
			}

			t.Fatalf("svg is not well formed: %v\n%s", err, svg)
		}
	}
}

func TestBarChart(t *testing.T) {
	svg := string(BarChart(chart.Chart{
		Frame:      ui.Frame{}.Size(ui.L560, ui.L320),
		XAxisTitle: "Quarter",
		YAxisTitle: "Orders",
	}).Series(testSeries).SVG())

	assertWellFormed(t, svg)

	if !strings.Contains(svg, `viewBox="0 0 560 320"`) {
		t.Errorf("expected the view box to follow the frame: %s", svg)
	}

	if got := strings.Count(svg, `<path d="M`); got != 8 {
		t.Errorf("expected 8 bars, got %d", got)
	}

	for _, want := range []string{">Quarter<", ">Orders<", ">Online<", ">Retail<", ">Q1<", ">Q4<", ">250<", ">160.5<", "Retail · Q4: 160.5"} {
		if !strings.Contains(svg, want) {
			t.Errorf("expected %q in %s", want, svg)
		}
	}

	if !strings.Contains(svg, `fill="#2a78d6"`) {
		t.Errorf("expected the default palette to be used")
	}
}

func TestBarChartRoundingAndColors(t *testing.T) {
	svg := string(BarChart(chart.Chart{
		Colors:        []ui.Color{ui.M0, "#ff0000/128"},
		LabelRounding: chart.RoundingRound,
	}).Series(testSeries).SVG())

	assertWellFormed(t, svg)

	if strings.Contains(svg, "160.5") {
		t.Errorf("expected the labels to be rounded: %s", svg)
	}

	if !strings.Contains(svg, `fill="var(--M0)"`) {
		t.Errorf("expected theme colors to be referenced as css variables: %s", svg)
	}

	if !strings.Contains(svg, `fill="color-mix(in srgb, #ff0000 50.2%, transparent)"`) {
		t.Errorf("expected the transparency to be applied: %s", svg)
	}
}

func TestBarChartStackedAndMarkers(t *testing.T) {
	markers := []Marker{
		{Label: "goal", SeriesIndex: 0, DataPointIndex: 1, Value: "200", Color: ui.SE0, Dashed: true},
		{Label: "dot", SeriesIndex: 1, DataPointIndex: 0, Value: "100", Color: "#000000", Round: true, Height: 8},
		{Label: "broken", SeriesIndex: 1, DataPointIndex: 0, Value: "nan"},
	}

	svg := string(BarChart(chart.Chart{}).Series(testSeries).Markers(markers).SVG())
	assertWellFormed(t, svg)

	if !strings.Contains(svg, `stroke-dasharray="3 3"`) || !strings.Contains(svg, ">goal: 200<") {
		t.Errorf("expected a dashed marker: %s", svg)
	}

	if !strings.Contains(svg, `r="4" fill="#000000"`) || !strings.Contains(svg, ">dot: 100<") {
		t.Errorf("expected a round marker: %s", svg)
	}

	if strings.Contains(svg, "broken") {
		t.Errorf("expected invalid markers to be ignored: %s", svg)
	}

	stacked := string(BarChart(chart.Chart{}).Series(testSeries).Markers(markers).Stacked(true).Horizontal(true).SVG())
	assertWellFormed(t, stacked)

	if strings.Contains(stacked, "goal") {
		t.Errorf("expected no markers on stacked charts: %s", stacked)
	}

	// the stack of Q4 is 370.5, so the axis must reach 400 and the total is labelled
	for _, want := range []string{">400<", ">370.5<"} {
		if !strings.Contains(stacked, want) {
			t.Errorf("expected %q in %s", want, stacked)
		}
	}
}

func TestNoData(t *testing.T) {
	svg := string(BarChart(chart.Chart{NoDataMessage: "nothing here"}).SVG())
	assertWellFormed(t, svg)

	if !strings.Contains(svg, ">nothing here<") || strings.Contains(svg, "<path") {
		t.Errorf("expected only the message: %s", svg)
	}

	empty := string(LineChart(chart.Chart{}).Series([]chart.Series{{Label: "x"}}).SVG())
	assertWellFormed(t, empty)
	if strings.Contains(empty, "<text") {
		t.Errorf("expected an empty image: %s", empty)
	}
}

func TestLineChart(t *testing.T) {
	series := []chart.Series{
		testSeries[0],
		{Label: "Area", Type: chart.ChartSeriesTypeArea, DataPoints: testSeries[1].DataPoints},
		{Label: "Bars", Type: chart.ChartSeriesTypeColumn, DataPoints: testSeries[1].DataPoints},
	}

	for _, curve := range []Curve{CurveStraight, CurveSmooth, CurveStepline} {
		svg := string(LineChart(chart.Chart{}).Series(series).Curve(curve).Markers(Markers{Size: 4, BorderColor: ui.M1}).ShowDataLabels(true).SVG())
		assertWellFormed(t, svg)

		if got := strings.Count(svg, `stroke-linejoin="round"`); got != 2 {
			t.Errorf("curve %d: expected 2 lines, got %d: %s", curve, got, svg)
		}

		if got := strings.Count(svg, `fill-opacity="0.15"`); got != 1 {
			t.Errorf("curve %d: expected 1 area, got %d", curve, got)
		}

		if got := strings.Count(svg, `<path d="M`); got != 2+1+4 {
			t.Errorf("curve %d: expected 2 lines, 1 area and 4 bars, got %d", curve, got)
		}

		if got := strings.Count(svg, `<circle`); got != 8+3 {
			t.Errorf("curve %d: expected 8 markers and 3 legend swatches, got %d", curve, got)
		}

		if !strings.Contains(svg, `stroke="var(--M1)"`) {
			t.Errorf("curve %d: expected the border color on the markers", curve)
		}

		if !strings.Contains(svg, ">210<") {
			t.Errorf("curve %d: expected data labels", curve)
		}
	}

	straight := string(LineChart(chart.Chart{}).Series(series[:1]).SVG())
	if strings.Contains(straight, "C") && strings.Contains(straight, "<circle") {
		t.Errorf("expected neither curves nor markers nor legend for a single straight series: %s", straight)
	}

	smooth := string(LineChart(chart.Chart{}).Series(series[:1]).Curve(CurveSmooth).SVG())
	if !strings.Contains(smooth, `d="M`) || !strings.Contains(smooth, "C") {
		t.Errorf("expected bezier curves: %s", smooth)
	}

	step := string(LineChart(chart.Chart{}).Series(series[:1]).Curve(CurveStepline).SVG())
	if !strings.Contains(step, "H") || !strings.Contains(step, "V") {
		t.Errorf("expected step lines: %s", step)
	}
}

func TestMonotoneTangents(t *testing.T) {
	points := []point{{x: 0, y: 0}, {x: 1, y: 10}, {x: 2, y: 10}, {x: 3, y: 0}}
	tangents := monotoneTangents(points)
	if tangents[1] != 0 || tangents[2] != 0 {
		t.Errorf("expected flat tangents at the plateau, got %v", tangents)
	}
}

func TestPieChart(t *testing.T) {
	series := []chart.Series{{Label: "Browser", DataPoints: []chart.DataPoint{
		{X: "Chrome", Y: 64}, {X: "Safari", Y: 19}, {X: "Firefox", Y: 9}, {X: "Other", Y: 8}, {X: "Negative", Y: -3},
	}}}

	svg := string(PieChart(chart.Chart{Frame: ui.Frame{}.Size(ui.L320, ui.L256)}).Series(series).SVG())
	assertWellFormed(t, svg)

	if got := strings.Count(svg, `<path d="M`); got != 4 {
		t.Errorf("expected 4 slices, got %d: %s", got, svg)
	}

	for _, want := range []string{">64%<", ">19%<", ">Chrome<", ">Other<", "Chrome: 64 (64%)"} {
		if !strings.Contains(svg, want) {
			t.Errorf("expected %q in %s", want, svg)
		}
	}

	if strings.Contains(svg, "Negative") {
		t.Errorf("expected negative values to be skipped")
	}

	donut := string(PieChart(chart.Chart{}).Series(series).ShowAsDonut(true).ShowAbsoluteValues(true).SVG())
	assertWellFormed(t, donut)
	if !strings.Contains(donut, ">64<") || strings.Contains(donut, "%<") {
		t.Errorf("expected absolute labels: %s", donut)
	}

	hidden := string(PieChart(chart.Chart{}).Series(series).ShowDataLabels(false).SVG())
	if strings.Contains(hidden, "%<") {
		t.Errorf("expected no labels: %s", hidden)
	}

	single := string(PieChart(chart.Chart{}).Series([]chart.Series{{DataPoints: []chart.DataPoint{{X: "all", Y: 1}}}}).ShowAsDonut(true).SVG())
	assertWellFormed(t, single)
	if !strings.Contains(single, ">100%<") {
		t.Errorf("expected a full ring: %s", single)
	}

	empty := string(PieChart(chart.Chart{NoDataMessage: "empty"}).Series([]chart.Series{{DataPoints: []chart.DataPoint{{X: "a", Y: 0}}}}).SVG())
	if !strings.Contains(empty, ">empty<") {
		t.Errorf("expected the no data message: %s", empty)
	}
}

func TestNiceScale(t *testing.T) {
	tests := []struct {
		lo, hi   float64
		min, max float64
		ticks    int
	}{
		{0, 210, 0, 250, 6},
		{90, 610, 0, 800, 5},
		{-400, 500, -400, 600, 6},
		{0, 0, 0, 1, 6},
		{0, 0.3, 0, 0.3, 4},
		{-5, 0, -5, 0, 6},
	}

	for _, tt := range tests {
		s := niceScale(tt.lo, tt.hi)
		if s.min != tt.min || s.max != tt.max || len(s.ticks) != tt.ticks {
			t.Errorf("niceScale(%v, %v) = %v..%v with %d ticks, want %v..%v with %d ticks", tt.lo, tt.hi, s.min, s.max, len(s.ticks), tt.min, tt.max, tt.ticks)
		}
	}

	if got := niceScale(0, 0.3).tickLabel(0.1); got != "0.1" {
		t.Errorf("expected a clean tick label, got %q", got)
	}
}

func TestCSSColor(t *testing.T) {
	tests := map[ui.Color]string{
		"":              "currentColor",
		"#112233":       "#112233",
		"#11223380":     "#11223380",
		"rgb(1,2,3)":    "rgb(1,2,3)",
		ui.M0:           "var(--M0)",
		ui.M0 + "/255":  "color-mix(in srgb, var(--M0) 100%, transparent)",
		"#112233/0":     "color-mix(in srgb, #112233 0%, transparent)",
		ui.A0 + "/junk": "var(--A0)",
	}

	for in, want := range tests {
		if got := cssColor(in); got != want {
			t.Errorf("cssColor(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFrameSize(t *testing.T) {
	tests := []struct {
		frame ui.Frame
		w, h  float64
	}{
		{ui.Frame{}, 560, 320},
		{ui.Frame{}.Size(ui.L560, ui.L320), 560, 320},
		{ui.Frame{Width: "300dp", Height: "100dp"}, 300, 100},
		{ui.Frame{Width: ui.Full}, 560, 320},
		{ui.Frame{Width: "280dp"}, 280, 160},
		{ui.Frame{MaxWidth: "280dp", Height: "200dp"}, 280, 200},
	}

	for _, tt := range tests {
		w, h := frameSize(tt.frame)
		if w != tt.w || h != tt.h {
			t.Errorf("frameSize(%v) = %v x %v, want %v x %v", tt.frame, w, h, tt.w, tt.h)
		}
	}
}

func TestEscaping(t *testing.T) {
	svg := string(BarChart(chart.Chart{XAxisTitle: `<script>alert("x")</script>`}).Series([]chart.Series{{Label: "a&b", DataPoints: []chart.DataPoint{{X: "<x>", Y: 1}}}}).SVG())
	assertWellFormed(t, svg)

	if strings.Contains(svg, "<script>") || strings.Contains(svg, "<x>") {
		t.Errorf("expected all text to be escaped: %s", svg)
	}
}
