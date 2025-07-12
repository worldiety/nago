// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"maps"
	"slices"
	"time"
)

type evt struct {
	v        core.View
	from, to time.Time
}

func timelineLane(iv ViewPort, evts ...evt) ui.THStack {
	return ui.HStack(
		ui.HStack().Frame(ui.Frame{Width: iv.LaneWidth.Length()}),
		ui.VStack(
			slices.Collect(func(yield func(view core.View) bool) {

				// background
				offset := 0.0
				colPart := 100 / float64(len(iv.Columns))
				for range iv.Columns {
					left := ui.Length(fmt.Sprintf("%f%%", offset))
					right := ui.Length(fmt.Sprintf("%f%%", 100-offset))
					yield(ui.HStack().Position(ui.Position{Type: ui.PositionAbsolute, Left: left, Right: right, Top: "0.5rem", Bottom: "0.5rem"}).Border(ui.Border{LeftWidth: ui.L1, LeftColor: ui.ColorIconsMuted}))
					offset += colPart
				}

				// actual events
				for _, e := range evts {
					yield(
						timelineEventPill(iv, e),
					)
				}

			})...,
		).Gap(ui.L4).Position(ui.Position{Type: ui.PositionOffset}).BackgroundColor(ui.ColorAccent).Frame(ui.Frame{Width: (100 - iv.LaneWidth).Length()}).Padding(ui.Padding{Top: ui.L16, Bottom: ui.L16}),
	).FullWidth()
}

func timelineEventPill(iv ViewPort, e evt) core.View {
	left := iv.Percent(e.from).Length()
	right := (100 - iv.Percent(e.to)).Length()
	return ui.HStack(
		ui.HStack(
			ui.HStack().BackgroundColor("#00ffff").Frame(ui.Frame{Width: ui.L12}),
			ui.HStack(e.v).BackgroundColor("#00ff00").FullWidth(),
		).
			Gap(ui.L2).
			Alignment(ui.Stretch).
			Border(ui.Border{}.Radius(ui.L8)).
			Frame(ui.Frame{MinHeight: ui.L40, Width: ui.Full}),
	).Position(ui.Position{ZIndex: 1}).
		Padding(ui.Padding{Left: left, Right: right}).
		Frame(ui.Frame{Width: ui.Full})
}

func renderTimelineYear(c TCalendar, ctx core.RenderContext) core.RenderNode {
	/*	const (
		widthLane     ui.Length = "16%"
		widthLaneRest ui.Length = "84%"
		widthMonth    ui.Length = "7%"
		rowHeight               = 16
	)*/

	widthLane := c.vp.LaneWidth.Length()
	widthMonth := ((100 - c.vp.LaneWidth) / Percent(len(c.vp.Columns))).Length()

	return ui.VStack(
		slices.Collect(func(yield func(core.View) bool) {

			yield(ui.HStack(
				slices.Collect(func(yield func(core.View) bool) {

					// header
					yield(ui.Text(c.vp.LaneLabel).BackgroundColor(ui.ColorCardTop).Frame(ui.Frame{Width: widthLane}))
					for _, col := range c.vp.Columns {
						yield(ui.Text(col.Label).BackgroundColor(ui.ColorCardTop).Frame(ui.Frame{Width: widthMonth}))
					}

				})...,
			).FullWidth())

			// test
			yield(
				timelineLane(c.vp,
					evt{
						v:    ui.Text("event in januar dessen name zu lang ist"),
						from: time.Date(c.vp.From.Year(), time.January, 1, 0, 0, 0, 0, time.UTC),
						to:   time.Date(c.vp.From.Year(), time.January, 31, 24, 0, 0, 0, time.UTC),
					},
					evt{
						v:    ui.Text("lololo"),
						from: time.Date(c.vp.From.Year(), time.July, 1, 0, 0, 0, 0, time.UTC),
						to:   time.Date(c.vp.From.Year(), time.September, 1, 0, 0, 0, 0, time.UTC),
					},

					evt{
						v:    ui.Text("cccc"),
						from: time.Date(c.vp.From.Year(), time.February, 1, 0, 0, 0, 0, time.UTC),
						to:   time.Date(c.vp.From.Year(), time.September, 1, 0, 0, 0, 0, time.UTC),
					},
				),
			)

			//lanes := mapLanes(c.vp, c.events)

		})...,
	).FullWidth().
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L16)).
		Render(ctx)
}

type laneCluster struct {
	Label  string
	Events []Event
}

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

	return res
}
