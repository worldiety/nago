// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"fmt"
	"maps"
	"slices"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// timelineLane renders a single lane within the timeline view.
//
// A lane consists of two main areas:
//  1. The **lane header** on the left, showing the label of the lane.
//  2. The **lane events area** on the right, which displays the lane's events,
//     background grid, and separators.
func timelineLane(colors Colors, iv ViewPort, cluster laneCluster) ui.THStack {
	// background padding look-through
	var bgTop ui.Length = "0rem"
	var bgBot ui.Length = "0rem"
	if cluster.first {
		bgBot = "0.2rem"
	} else {
		bgTop = "0.2rem"
	}

	if !cluster.last {
		bgBot = "0.0rem"
	}

	return ui.HStack(
		// lane header
		ui.HStack(
			ui.VStack(
				ui.Text(cluster.Label).TextAlignment(ui.TextAlignCenter)).
				BackgroundColor(colors.LaneBackground).
				FullWidth(),
		).Alignment(ui.Stretch).Padding(ui.Padding{Top: bgTop, Bottom: bgBot}).Frame(ui.Frame{Width: iv.LaneWidth.Length()}),

		// lane events
		ui.VStack(
			slices.Collect(func(yield func(view core.View) bool) {

				// background

				yield(ui.HStack().Position(ui.Position{Type: ui.PositionAbsolute, Left: "0rem", Right: "0rem", Top: bgTop, Bottom: bgBot}).BackgroundColor(colors.LaneBackground))

				offset := 0.0
				colPart := 100 / float64(len(iv.Columns))
				for range iv.Columns {
					left := ui.Length(fmt.Sprintf("%f%%", offset))
					right := ui.Length(fmt.Sprintf("%f%%", 100-offset))
					yield(ui.HStack().Position(ui.Position{Type: ui.PositionAbsolute, Left: left, Right: right, Top: "0.5rem", Bottom: "0.5rem"}).Border(ui.Border{LeftWidth: ui.L1, LeftColor: ui.ColorIconsMuted}))
					offset += colPart
				}

				// actual events
				for _, e := range cluster.Events {
					yield(
						timelineEventPill(colors, iv, e),
					)
				}

			})...,
		).Gap(ui.L4).Position(ui.Position{Type: ui.PositionOffset}).Frame(ui.Frame{Width: (100 - iv.LaneWidth).Length()}).Padding(ui.Padding{Top: ui.L16, Bottom: ui.L16}),
	).FullWidth().Alignment(ui.Stretch)
}

// timelineEventPill renders a timeline event as a pill with optional pre/post durations,
// a category color bar, and an event body that can include hover/click actions.
func timelineEventPill(colors Colors, iv ViewPort, e Event) core.View {
	left := iv.Percent(e.From.At).Length()
	right := (100 - iv.Percent(e.To.At)).Length()
	return ui.HStack(
		// pre-duration
		ui.If(e.From.Offset.Duration > 0, ui.HStack(
			ui.ImageIcon(e.From.Offset.Icon).FillColor(colors.PrePostForeground).AccessibilityLabel(e.From.Offset.Label),
		).BackgroundColor(colors.PrePostBackground).Border(ui.Border{}.Radius(ui.L8)).Padding(ui.Padding{}.Horizontal(ui.L16))),

		ui.HStack(
			// category color
			ui.HStack().BackgroundColor(e.Category.Color).Frame(ui.Frame{MinWidth: ui.L12}).AccessibilityLabel(e.Category.Label),
			ui.HStack(ui.Text(e.Label).Color(colors.EventText).AccessibilityLabel(e.Label)).BackgroundColor(colors.EventBackground).
				FullWidth().
				With(func(stack ui.THStack) ui.THStack {
					if e.Action != nil {
						stack = stack.HoveredBackgroundColor(colors.EventHoverBackground).Action(e.Action)
					}

					return stack
				}).Padding(ui.Padding{}.All(ui.L8)),
		).
			Gap(ui.L2).
			Alignment(ui.Stretch).
			Border(ui.Border{}.Radius(ui.L8)).
			Frame(ui.Frame{MinHeight: ui.L40, Width: ui.Full}),

		// post-duration
		ui.If(e.To.Offset.Duration > 0, ui.HStack(
			ui.ImageIcon(e.To.Offset.Icon).FillColor(colors.PrePostForeground).AccessibilityLabel(e.To.Offset.Label),
		).BackgroundColor(colors.PrePostBackground).Border(ui.Border{}.Radius(ui.L8)).Padding(ui.Padding{}.Horizontal(ui.L16))),
	).Position(ui.Position{ZIndex: 1}).
		Alignment(ui.Stretch).
		Padding(ui.Padding{Left: left, Right: right}).
		Frame(ui.Frame{Width: ui.Full})
}

// renderTimelineYear builds a year-view calendar layout with a header row
// (lane label + month labels) and corresponding timeline lanes for events.
func renderTimelineYear(c TCalendar, ctx core.RenderContext) core.RenderNode {

	widthLane := c.vp.LaneWidth.Length()
	widthMonth := ((100 - c.vp.LaneWidth) / Percent(len(c.vp.Columns))).Length()

	return ui.VStack(
		slices.Collect(func(yield func(core.View) bool) {

			yield(ui.HStack(
				slices.Collect(func(yield func(core.View) bool) {

					// header
					yield(ui.Text(c.vp.LaneLabel).BackgroundColor(c.colors.Header).Frame(ui.Frame{Width: widthLane}).Padding(ui.Padding{}.All(ui.L16)))
					for _, col := range c.vp.Columns {
						yield(ui.Text(col.Label).BackgroundColor(c.colors.Header).Frame(ui.Frame{Width: widthMonth}).Padding(ui.Padding{}.Vertical(ui.L16)))
					}

				})...,
			).FullWidth())

			lanes := mapLanes(c.vp, c.events)
			for _, lane := range lanes {
				yield(timelineLane(c.colors, c.vp, lane))
			}

		})...,
	).FullWidth().
		Border(ui.Border{}.Radius(ui.L16)).
		Render(ctx)
}

// laneCluster represents a group of events within the same lane,
// including metadata to mark the first and last lane in the timeline.
type laneCluster struct {
	Label  string
	Events []Event
	first  bool
	last   bool
}

// mapLanes groups events by their lane label within the viewport range,
// clamps their start/end times to fit, and marks the first/last lane.
func mapLanes(vp ViewPort, events []Event) []laneCluster {
	tmp := map[string]laneCluster{}
	for _, event := range events {
		// ignore, if out of range in the future
		if event.From.At.After(vp.To) {
			continue
		}

		// ignore, if out of range in the past
		if event.To.At.Before(vp.From) {
			continue
		}

		// clamp start
		if event.From.At.Before(vp.From) {
			event.From.At = vp.From
		}

		// clamp end
		if event.To.At.After(vp.To) {
			event.To.At = vp.To
		}

		l := tmp[event.Lane.Label]
		l.Label = event.Lane.Label
		l.Events = append(l.Events, event)
		tmp[event.Lane.Label] = l
	}

	res := make([]laneCluster, 0, len(tmp))
	for _, k := range slices.Sorted(maps.Keys(tmp)) {
		res = append(res, tmp[k])
	}

	if len(res) > 0 {
		res[0].first = true
		res[len(res)-1].last = true
	}

	return res
}
