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
	"math"
	"slices"
	"strconv"
	"time"
)

type evt struct {
	v        core.View
	from, to time.Time
}

func timelineLane(a, b ui.Length, year int, evts ...evt) ui.THStack {
	return ui.HStack(
		ui.HStack().Frame(ui.Frame{Width: a}),
		ui.VStack(
			slices.Collect(func(yield func(view core.View) bool) {

				// background
				offset := 0.0
				for range time.December {
					left := ui.Length(fmt.Sprintf("%f%%", offset))
					right := ui.Length(fmt.Sprintf("%f%%", 100-offset))
					yield(ui.HStack().Position(ui.Position{Type: ui.PositionAbsolute, Left: left, Right: right, Top: "0%", Bottom: "0%"}).Border(ui.Border{LeftWidth: ui.L1, LeftColor: ui.ColorIconsMuted}))
					offset += 8.3
				}

				// actual events
				for _, e := range evts {
					yield(
						timelineEventPill(year, e),
					)
				}

			})...,
		).Gap(ui.L4).Position(ui.Position{Type: ui.PositionOffset}).BackgroundColor(ui.ColorAccent).Frame(ui.Frame{Width: b}).Padding(ui.Padding{Top: ui.L8, Bottom: ui.L8}),
	).FullWidth()
}

func timelineEventPill(year int, e evt) core.View {
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
		Padding(ui.Padding{Left: cssAbsLeftPercentInYear(year, e.from), Right: cssAbsRightPercentInYear(year, e.to)}).
		Frame(ui.Frame{Width: ui.Full})
}

func renderTimelineYear(c TCalendar, ctx core.RenderContext) core.RenderNode {
	const (
		widthLane     ui.Length = "16%"
		widthLaneRest ui.Length = "84%"
		widthMonth    ui.Length = "7%"
		rowHeight               = 16
	)

	return ui.VStack(
		slices.Collect(func(yield func(core.View) bool) {

			yield(ui.HStack(
				slices.Collect(func(yield func(core.View) bool) {

					// header
					yield(ui.Text(strconv.Itoa(c.year)).BackgroundColor(ui.ColorCardTop).Frame(ui.Frame{Width: widthLane}))
					for m := range time.December {
						yield(ui.Text((m + 1).String()).BackgroundColor(ui.ColorCardTop).Frame(ui.Frame{Width: widthMonth}))
					}

				})...,
			).FullWidth())

			// test
			yield(
				timelineLane(widthLane, widthLaneRest, c.year,
					evt{
						v:    ui.Text("event in januar dessen name zu lang ist"),
						from: time.Date(c.year, time.January, 1, 0, 0, 0, 0, time.UTC),
						to:   time.Date(c.year, time.January, 31, 24, 0, 0, 0, time.UTC),
					},
					evt{
						v:    ui.Text("lololo"),
						from: time.Date(c.year, time.July, 1, 0, 0, 0, 0, time.UTC),
						to:   time.Date(c.year, time.September, 1, 0, 0, 0, 0, time.UTC),
					},

					evt{
						v:    ui.Text("cccc"),
						from: time.Date(c.year, time.February, 1, 0, 0, 0, 0, time.UTC),
						to:   time.Date(c.year, time.September, 1, 0, 0, 0, 0, time.UTC),
					},
				),
			)

			// some bubbles
			lanes := mapLanes(c.year, c.events)
			for _, lane := range lanes {
				yield(ui.HStack(
					slices.Collect(func(yield func(core.View) bool) {

						// background
						yield(ui.Text(lane.Label).Frame(ui.Frame{Width: widthLane}))
						for range time.December {
							yield(ui.HStack().Frame(ui.Frame{Width: widthMonth}).Border(ui.Border{LeftWidth: ui.L1, LeftColor: ui.ColorIconsMuted}))
						}

						yield(ui.HStack(

							slices.Collect(func(yield func(core.View) bool) {
								// absolute events
								top := float64(0)
								for _, event := range lane.Events {

									yield(
										ui.VStack(ui.Text(event.Label)).
											BackgroundColor(ui.ColorError).
											Position(ui.Position{Type: ui.PositionAbsolute, Top: ui.L(top), Left: cssAbsLeftPercentInYear(c.year, event.From.At), Right: cssAbsRightPercentInYear(c.year, event.To.At)}),
									)

									top += rowHeight
								}

							})...,
						).BackgroundColor("#330000aa").Position(ui.Position{Type: ui.PositionAbsolute, Top: "0px", Bottom: "0px", Left: widthLane, Right: "0rem"}).Frame(ui.Frame{Height: ui.L160}))

					})...,
				).Alignment(ui.Stretch).Position(ui.Position{Type: ui.PositionOffset}).Frame(ui.Frame{Width: ui.Full, Height: ui.L(float64(len(lane.Events)) * rowHeight)}))
			}

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

func mapLanes(year int, events []Event) []laneCluster {
	tmp := map[string]laneCluster{}
	for _, event := range events {
		if event.From.At.Year() > year {
			continue
		}

		if event.To.At.Year() < year {
			continue
		}

		if event.From.At.Year() < year {
			event.From.At = time.Date(year, time.January, 0, 0, 0, 0, 0, time.Local)
		}

		if event.To.At.Year() > year {
			event.From.At = time.Date(year, time.December, 31, 23, 59, 59, 0, time.Local)
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

func percentInYear(year int, t time.Time) float64 {
	start := float64(time.Date(year, time.January, 1, 0, 0, 0, 0, t.Location()).Unix())
	end := float64(time.Date(year, time.December, 31, 24, 0, 0, -1, t.Location()).Unix())
	now := float64(t.Unix())
	p := (now - start) / (end - start) * 100
	p = math.Round(p*100) / 100
	return p
}

func cssPercentInYear(year int, t time.Time) ui.Length {
	p := percentInYear(year, t)
	return ui.Length(strconv.FormatFloat(p, 'f', -1, 64)) + "%"
}

func cssAbsLeftPercentInYear(year int, t time.Time) ui.Length {
	p := percentInYear(year, t)
	return ui.Length(strconv.FormatFloat(p, 'f', -1, 64)) + "%"
}

func cssAbsRightPercentInYear(year int, t time.Time) ui.Length {
	p := percentInYear(year, t)
	return ui.Length(strconv.FormatFloat(100-p, 'f', -1, 64)) + "%"
}
