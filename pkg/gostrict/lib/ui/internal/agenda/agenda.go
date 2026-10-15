// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package agenda is a program, which renders a calendar, pickers for dates, times and time frames and a flowchart, for
// the test of the package.
package agenda

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

// Render renders the window.
func Render(wnd ui.Window) ui.View {
	day := ui.DateStateOf(wnd, "day").Init(func() ui.Date { return ui.Date{Day: 24, Month: 12, Year: 2025} })
	frame := ui.TimeFrameStateOf(wnd, "frame").Init(func() ui.TimeFrame {
		return ui.TimeFrame{StartTime: ui.TimeOf(2025, 12, 24, 10, 0, 0), EndTime: ui.TimeOf(2025, 12, 24, 11, 30, 0), Timezone: "Europe/Berlin"}
	})
	length := ui.DurationStateOf(wnd, "length").Init(func() ui.Duration { return 90 * ui.Minute })
	model := ui.FlowchartModelStateOf(wnd, "model").Init(func() ui.FlowchartModel {
		return ui.FlowchartModel{
			Nodes: list.Of(
				ui.FlowchartNode{ID: "a", Type: ui.FlowchartNodeTypeStart, Label: "Start"},
				ui.FlowchartNode{ID: "b", Position: ui.FlowchartPoint{X: 200, Y: 50}, Label: "Ende", Type: ui.FlowchartNodeTypeEnd},
			),
			Edges: list.Of(ui.FlowchartEdge{ID: "ab", SourceNodeID: "a", TargetNodeID: "b", MarkerEnd: ui.FlowchartEdgeMarkerArrow}),
		}
	})
	action := ui.FlowchartActionStateOf(wnd, "action")

	start := ui.TimeOf(2025, 12, 24, 9, 0, 0)
	events := ui.Calendar(
		ui.CalendarEvent("Bescherung", start, start.Add(2*ui.Hour)).
			Categories(ui.CalendarCategory{Color: "#ff0000", Label: "Familie"}).
			Organiser("Anna").
			Location("Zuhause"),
		ui.CalendarEvent("Spaziergang", start.Add(3*ui.Hour), start.Add(4*ui.Hour+30*ui.Minute)).
			Cancelled(true).
			Chips(ui.CalendarChip{Label: "Draußen"}),
	).ViewPort(ui.CalendarDay(2025, 12, 24)).Style(ui.CalendarStartTimeSequence).FullWidth()

	label := "Tag: " + day.Get().String() + ", " + start.Format("15:04") + ", Dauer " + frame.Get().StartTime.Date().String()
	return ui.ScrollView(
		ui.VStack(
			ui.WindowTitle("Agenda"),
			ui.H2("Kalender"),
			events,
			ui.Text(label),
			ui.SingleDatePicker("Tag", day.Get(), day).SupportingText("Wann?"),
			ui.TimeframePicker("Zeitraum", frame).Title("Zeitraum"),
			ui.TimepickerPicker("Dauer", length).Hours(true).Minutes(true),
			ui.FlowChart(model.Get()).
				InputValue(model).
				ActionValue(action).
				Toolbar(ui.FlowchartToolbar{Actions: list.Of(ui.FlowChartToolbarActionAutoLayout)}).
				Background(ui.FlowchartBackground{GridGap: 16}).
				Frame(ui.Frame{Height: ui.L320}),
		),
	)
}
