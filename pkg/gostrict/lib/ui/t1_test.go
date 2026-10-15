// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/calendar"
	"go.wdy.de/nago/presentation/ui/flowchart"
	"go.wdy.de/nago/presentation/ui/timeframe"
)

// TestEnumsT1 compares the constants of the enums of batch t1 with the constants of nago.
func TestEnumsT1(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"CalendarTimelineYear", int64(CalendarTimelineYear), int64(calendar.TimelineYear)},
		{"CalendarTimelineDay", int64(CalendarTimelineDay), int64(calendar.TimelineDay)},
		{"CalendarStartTimeSequence", int64(CalendarStartTimeSequence), int64(calendar.StartTimeSequence)},
		{"CalendarPillTimeYYYYMMDD", int64(CalendarPillTimeYYYYMMDD), int64(calendar.PillTimeYYYYMMDD)},
		{"CalendarPillTimeHHMM", int64(CalendarPillTimeHHMM), int64(calendar.PillTimeHHMM)},
		{"CalendarPillTimeNone", int64(CalendarPillTimeNone), int64(calendar.PillTimeNone)},
		{"TimeframeClassicFormat", int64(TimeframeClassicFormat), int64(timeframe.ClassicFormat)},
		{"FlowChartLayoutHorizontal", int64(FlowChartLayoutHorizontal), int64(flowchart.FlowChartLayoutHorizontal)},
		{"FlowChartLayoutVertical", int64(FlowChartLayoutVertical), int64(flowchart.FlowChartLayoutVertical)},
		{"FlowChartBackgroundGridStyleDots", int64(FlowChartBackgroundGridStyleDots), int64(flowchart.FlowChartBackgroundGridStyleDots)},
		{"FlowChartBackgroundGridStyleLines", int64(FlowChartBackgroundGridStyleLines), int64(flowchart.FlowChartBackgroundGridStyleLines)},
		{"FlowchartEdgeStyleSolid", int64(FlowchartEdgeStyleSolid), int64(flowchart.EdgeStyleSolid)},
		{"FlowchartEdgeStyleDashed", int64(FlowchartEdgeStyleDashed), int64(flowchart.EdgeStyleDashed)},
		{"FlowchartEdgeStyleDotted", int64(FlowchartEdgeStyleDotted), int64(flowchart.EdgeStyleDotted)},
		{"FlowchartEdgeMarkerNone", int64(FlowchartEdgeMarkerNone), int64(flowchart.EdgeMarkerNone)},
		{"FlowchartEdgeMarkerArrow", int64(FlowchartEdgeMarkerArrow), int64(flowchart.EdgeMarkerArrow)},
		{"FlowchartNodeTypeDefault", int64(FlowchartNodeTypeDefault), int64(flowchart.NodeTypeDefault)},
		{"FlowchartNodeTypeStart", int64(FlowchartNodeTypeStart), int64(flowchart.NodeTypeStart)},
		{"FlowchartNodeTypeEnd", int64(FlowchartNodeTypeEnd), int64(flowchart.NodeTypeEnd)},
		{"FlowchartNodeStyleDefault", int64(FlowchartNodeStyleDefault), int64(flowchart.NodeStyleDefault)},
		{"FlowchartNodeStyleNone", int64(FlowchartNodeStyleNone), int64(flowchart.NodeStyleNone)},
		{"FlowChartToolbarPositionTopRight", int64(FlowChartToolbarPositionTopRight), int64(flowchart.FlowChartToolbarPositionTopRight)},
		{"FlowChartToolbarPositionCenterRight", int64(FlowChartToolbarPositionCenterRight), int64(flowchart.FlowChartToolbarPositionCenterRight)},
		{"FlowChartToolbarPositionBottomRight", int64(FlowChartToolbarPositionBottomRight), int64(flowchart.FlowChartToolbarPositionBottomRight)},
		{"FlowChartToolbarPositionBottomCenter", int64(FlowChartToolbarPositionBottomCenter), int64(flowchart.FlowChartToolbarPositionBottomCenter)},
		{"FlowChartToolbarPositionBottomLeft", int64(FlowChartToolbarPositionBottomLeft), int64(flowchart.FlowChartToolbarPositionBottomLeft)},
		{"FlowChartToolbarPositionCenterLeft", int64(FlowChartToolbarPositionCenterLeft), int64(flowchart.FlowChartToolbarPositionCenterLeft)},
		{"FlowChartToolbarPositionTopLeft", int64(FlowChartToolbarPositionTopLeft), int64(flowchart.FlowChartToolbarPositionTopLeft)},
		{"FlowChartToolbarPositionTopCenter", int64(FlowChartToolbarPositionTopCenter), int64(flowchart.FlowChartToolbarPositionTopCenter)},
		{"FlowChartToolbarOrientationHorizontal", int64(FlowChartToolbarOrientationHorizontal), int64(flowchart.FlowChartToolbarOrientationHorizontal)},
		{"FlowChartToolbarOrientationVertical", int64(FlowChartToolbarOrientationVertical), int64(flowchart.FlowChartToolbarOrientationVertical)},
		{"FlowChartToolbarActionAutoLayout", int64(FlowChartToolbarActionAutoLayout), int64(flowchart.FlowChartToolbarActionAutoLayout)},
		{"Millisecond", int64(Millisecond), int64(time.Millisecond / time.Millisecond)},
		{"Second", int64(Second), int64(time.Second / time.Millisecond)},
		{"Minute", int64(Minute), int64(time.Minute / time.Millisecond)},
		{"Hour", int64(Hour), int64(time.Hour / time.Millisecond)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s: got %d, want %d", e.name, e.got, e.want)
		}
	}
}

// TestTimeT1 checks the conversions of times and dates.
func TestTimeT1(t *testing.T) {
	tm := TimeOf(2025, 12, 24, 18, 30, 15)
	want := time.Date(2025, 12, 24, 18, 30, 15, 0, time.Local)
	if tm.nago() != want || int64(tm) != want.UnixMilli() {
		t.Fatalf("TimeOf: %v", tm.nago())
	}
	if tm.Year() != 2025 || tm.Month() != 12 || tm.Day() != 24 || tm.Hour() != 18 || tm.Minute() != 30 {
		t.Error("fields")
	}
	if got := tm.Format("2006-01-02 15:04"); got != "2025-12-24 18:30" {
		t.Error(got)
	}
	d := tm.Date()
	if d != (Date{Day: 24, Month: 12, Year: 2025}) {
		t.Fatalf("Date: %v", d)
	}
	if d.Time() != TimeOf(2025, 12, 24, 0, 0, 0) {
		t.Error("Date.Time")
	}
	if d.String() != "2025.12.24" || d.IsZero() || !(Date{}).IsZero() {
		t.Error("Date.String or IsZero")
	}
	if tm.Add(90*Minute) != TimeOf(2025, 12, 24, 20, 0, 15) {
		t.Error("Add")
	}
	if dateFrom(d.nago()) != d {
		t.Error("date round trip")
	}
	if durationFrom(Duration(90*Minute).nago()) != 90*Minute || Duration(90*Minute).nago() != 90*time.Minute {
		t.Error("duration")
	}
	tf := TimeFrame{StartTime: tm, EndTime: tm.Add(Hour), Timezone: "Europe/Berlin"}
	if timeFrameFrom(tf.nago()) != tf {
		t.Error("time frame round trip")
	}
	if tf.nago().Timezone != xtime.Timezone("Europe/Berlin") {
		t.Error("timezone")
	}
	if Now() <= 0 {
		t.Error("Now")
	}
}

// TestViewsT1 builds each new view natively through the mirror functions.
func TestViewsT1(t *testing.T) {
	ev := CalendarEvent("a", TimeOf(2025, 1, 2, 3, 0, 0), TimeOf(2025, 1, 2, 4, 0, 0)).
		Action(func() {}).
		FromOffset(CalendarOffset{Duration: 1000, Label: "bus"}).
		ToOffset(CalendarOffset{Label: "train"}).
		Categories(CalendarCategory{Color: "#ff0000", Label: "x"}).
		Pillar(CalendarPillar{Label: "p"}).
		Lane(CalendarLane{Label: "l"}).
		Chips(CalendarChip{Label: "c"}).
		Organiser("o").Location("l").Cancelled(true).
		Render(func(style CalendarStyle) View { return Text("custom") })
	vp := CalendarYear(2025).From(TimeOf(2025, 1, 1, 0, 0, 0)).To(TimeOf(2025, 12, 31, 0, 0, 0)).Columns(CalendarColumn{Label: "a"}).LaneWidth(10).TimeStyle(CalendarPillTimeNone)
	if p := CalendarDay(2025, 1, 2).Percent(TimeOf(2025, 1, 2, 12, 0, 0)); p < 49.9 || p > 50.1 {
		t.Errorf("Percent: %v", p)
	}
	cal := Calendar(ev).Append(ev).ViewPort(vp).Style(CalendarTimelineYear).Colors(CalendarDefaultColors()).MaxCategories(2).FullWidth().Frame(Frame{})
	flow := FlowChart(FlowchartModel{Nodes: list.Of(FlowchartNode{ID: "a"})}).
		Toolbar(FlowchartToolbar{Actions: list.Of(FlowChartToolbarActionAutoLayout)}).
		Menu(FlowchartMenu(FlowchartPoint{X: 1}, Text("menu"))).
		CustomContents(FlowchartCustomContent("a", Text("content")).Menu(Text("m"))).
		AppendCustomContent(FlowchartCustomContent("b", Text("content"))).
		Background(FlowchartBackground{GridGap: 8}).Layout(FlowChartLayoutVertical).
		NodesDraggable(true).NodesConnectable(true).EdgesEditable(true).ElementsSelectable(true).
		MinZoom(0.5).MaxZoom(2).ReadOnly(false).Movable(true).FullWidth().Frame(Frame{}).WithFrame(func(f Frame) Frame { return f })
	m := FlowchartModel{}.WithNodes(FlowchartNode{ID: "a"}).AppendNodes(FlowchartNode{ID: "b"}).WithEdges(FlowchartEdge{ID: "e"}).AppendEdges(FlowchartEdge{ID: "f"})
	if m.Nodes.Len() != 2 || m.Edges.Len() != 2 || flowchartModelFrom(m.nago()).Nodes.Len() != 2 {
		t.Error("model")
	}
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.gostrict.t1")
		cfg.RootView("t1", func(wnd core.Window) core.View {
			w := WindowOf(wnd)
			frame := TimeFrameStateOf(w, "frame")
			day := DateStateOf(w, "day")
			end := DateStateOf(w, "end")
			length := DurationStateOf(w, "length")
			views := map[string]View{
				"calendar":   cal,
				"flowchart":  flow.InputValue(FlowchartModelStateOf(w, "m")).ActionValue(FlowchartActionStateOf(w, "a")),
				"timeframe":  TimeframePicker("t", frame).Disabled(true).Title("t").SupportingText("s").ErrorText("e").Format(TimeframeClassicFormat).Padding(Padding{}).Frame(Frame{}).WithFrame(func(f Frame) Frame { return f }).Border(Border{}).Visible(true).AccessibilityLabel("a"),
				"timepicker": TimepickerPicker("t", length).Disabled(false).Title("t").SupportingText("s").ErrorText("e").Format(TimepickerDecomposedFormat).Hours(true).Minutes(true).Days(true).Seconds(true).Padding(Padding{}).Frame(Frame{}).WithFrame(func(f Frame) Frame { return f }).Border(Border{}).Visible(true).AccessibilityLabel("a"),
				"single":     SingleDatePicker("d", Date{Day: 1, Month: 2, Year: 2025}, day).Disabled(false).SupportingText("s").ErrorText("e").DoubleMode(true).Optional(true).Padding(Padding{}).Frame(Frame{}).WithFrame(func(f Frame) Frame { return f }).Border(Border{}).Visible(true).AccessibilityLabel("a"),
				"range":      RangeDatePicker("d", Date{}, day, Date{}, end),
			}
			for name, v := range views {
				if Core(v) == nil {
					t.Errorf("%s renders no view", name)
				}
			}
			return Core(Text("ok"))
		})
	})
	app.Open(t, nil, "t1")
}
