// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"maps"
	"slices"
	"strconv"

	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/solid"
	"go.wdy.de/nago/presentation/ui"
)

func renderStartTimeSequence(c TCalendar, ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		// card top
		ui.VStack(
			ui.Text(c.vp.LaneLabel).Font(ui.BodyLarge),
			ui.HLineWithColor(ui.ColorAccent).Padding(ui.Padding{}),
		).Alignment(ui.Stretch).
			Padding(ui.Padding{}.All(ui.L16)),

		ui.VStack(
			// card body
			ui.ForEach(mapStartTimeClusterCalWeeks(c.vp, c.events), func(t startTimeCluster) core.View {

				// cluster with common start label
				return ui.HStack(
					// header
					ui.HStack(
						ui.ImageIcon(icons.Clock),
						ui.Text("Woche "+t.week.String()),
					).BackgroundColor(ui.ColorContainer).
						Gap(ui.L4).
						Padding(ui.Padding{}.All(ui.L16)).
						Border(ui.Border{}.Radius(ui.L8)).
						Frame(ui.Frame{Width: ui.L200}),

					// events
					ui.VStack(
						ui.ForEach(t.events, func(evt Event) core.View {
							if evt.Render != nil {
								return evt.Render(c.style)
							}

							return startTimeSeqPill(c, evt, t.isDayOnly)
						})...,
					).Alignment(ui.Stretch).
						Gap(ui.L4).
						FullWidth(),
				).FullWidth().
					Gap(ui.L16).
					Alignment(ui.Stretch)
			})...,
		).FullWidth().
			Gap(ui.L16).
			BackgroundColor(ui.ColorCardBody).
			Padding(ui.Padding{}.All(ui.L16)),
	).
		Alignment(ui.Leading).
		BackgroundColor(ui.ColorCardTop).
		Border(ui.Border{}.Radius(ui.L16)).
		Frame(c.frame).Render(ctx)
}

func startTimeSeqPill(c TCalendar, evt Event, timeOnly bool) core.View {
	colors := c.colors

	isTimePoint := evt.To.At.Equal(evt.From.At)

	var timeStr string
	if isTimePoint {
		if timeOnly {
			timeStr = evt.From.At.Format("15:04")
		} else {
			if evt.To.At.Hour() == 0 && evt.To.At.Minute() == 0 && evt.From.At.Hour() == 0 && evt.To.At.Minute() == 0 {
				timeStr = evt.From.At.Format(xtime.GermanDate)
			} else {
				timeStr = evt.From.At.Format(xtime.GermanDateTime)
			}
		}
	} else {
		if timeOnly {
			timeStr = evt.From.At.Format("15:04") + " - " + evt.To.At.Format("15:04")
		} else {
			if evt.To.At.Hour() == 0 && evt.To.At.Minute() == 0 && evt.From.At.Hour() == 0 && evt.To.At.Minute() == 0 {
				timeStr = evt.From.At.Format(xtime.GermanDate) + " - " + evt.To.At.Format(xtime.GermanDate)
			} else {
				timeStr = evt.From.At.Format(xtime.GermanDateTime) + " - " + evt.To.At.Format(xtime.GermanDateTime)
			}
		}
	}

	return ui.HStack(
		// category color
		ui.HStack().BackgroundColor(evt.Category.Color).Frame(ui.Frame{MinWidth: ui.L12}).AccessibilityLabel(evt.Category.Label),
		ui.VStack(
			ui.Text(evt.Label).Font(ui.BodyLarge),
			ui.HStack(
				ui.ImageIcon(icons.Clock),
				ui.Text(timeStr),
			).Gap(ui.L4),
			ui.If(evt.Organiser != "", ui.HStack(
				ui.ImageIcon(icons.User),
				ui.Text(evt.Organiser),
			).Gap(ui.L4)),
			ui.If(evt.Location != "", ui.HStack(
				ui.ImageIcon(icons.MapPinAlt),
				ui.Text(evt.Location),
			).Gap(ui.L4)),
		).
			BackgroundColor(colors.EventBackground).
			Alignment(ui.Leading).
			Gap(ui.L4).
			TextColor(colors.EventText).
			FullWidth().
			With(func(stack ui.TVStack) ui.TVStack {
				if evt.Action != nil {
					stack = stack.HoveredBackgroundColor(colors.EventHoverBackground).Action(evt.Action)
				}

				return stack
			}).Padding(ui.Padding{}.All(ui.L8)),
	).
		Gap(ui.L2).
		Alignment(ui.Stretch).
		Border(ui.Border{}.Radius(ui.L8)).
		Frame(ui.Frame{MinHeight: ui.L40, Width: ui.Full, Height: ui.Full})
}

type startTimeCluster struct {
	week      isoCalWeek
	events    []Event
	first     bool
	last      bool
	isDayOnly bool
}

type isoCalWeek struct {
	year int
	week int
}

func (i isoCalWeek) String() string {
	return strconv.Itoa(i.week)
}

func mapStartTimeClusterCalWeeks(vp ViewPort, events []Event) []startTimeCluster {
	tmp := map[isoCalWeek]startTimeCluster{}
	for _, event := range events {
		// ignore, if out of range in the future
		if event.From.At.After(vp.To) {
			continue
		}

		// ignore, if out of range in the past
		if event.To.At.Before(vp.From) {
			continue
		}

		y, w := event.From.At.ISOWeek()
		iso := isoCalWeek{y, w}
		l := tmp[iso]
		l.week = iso
		l.events = append(l.events, event)
		slices.SortFunc(l.events, func(a, b Event) int {
			return a.From.At.Compare(b.From.At)
		})
		tmp[iso] = l
	}

	res := make([]startTimeCluster, 0, len(tmp))
	for _, k := range slices.SortedFunc(maps.Keys(tmp), func(week isoCalWeek, week2 isoCalWeek) int {
		if week.year != week2.year {
			return week.year - week2.year
		}

		return week.week - week2.week
	}) {
		res = append(res, tmp[k])
	}

	if len(res) > 0 {
		res[0].first = true
		res[len(res)-1].last = true
	}

	return res
}
