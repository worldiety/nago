// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/solid"
	"go.wdy.de/nago/presentation/ui"
)

type cluster interface {
	String(bnd i18n.Bundler) string
	Events() []Event
	TimeStyle() seqPillTimeHint
}

func renderStartTimeSequence(c TCalendar, ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		// card top
		ui.VStack(
			ui.Text(c.vp.LaneLabel(ctx.Window())).Font(ui.BodyLarge),
			ui.HLineWithColor(ui.ColorAccent).Padding(ui.Padding{}),
		).Alignment(ui.Stretch).
			Padding(ui.Padding{}.All(ui.L16)),

		ui.VStack(
			// card body
			ui.ForEach(c.mapStartTimeCluster(c.vp, c.events), func(t cluster) core.View {
				tWidth := ui.Full
				isLarge := ctx.Window().Info().SizeClass >= core.SizeClassLarge
				if isLarge {
					tWidth = ui.L200
				}

				views := []core.View{
					// header
					ui.HStack(
						ui.ImageIcon(icons.Clock),
						ui.Text(t.String(ctx.Window())),
					).BackgroundColor(ui.ColorContainer).
						Gap(ui.L4).
						Padding(ui.Padding{}.All(ui.L16)).
						Border(ui.Border{}.Radius(ui.L8)).
						Frame(ui.Frame{Width: tWidth}),

					// events
					ui.VStack(
						ui.ForEach(t.Events(), func(evt Event) core.View {
							if evt.Render != nil {
								return evt.Render(c.style)
							}

							return startTimeSeqPill(c, evt, t.TimeStyle(), isLarge)
						})...,
					).Alignment(ui.Stretch).
						Gap(ui.L4).
						FullWidth(),
				}

				// cluster with common start label
				if ctx.Window().Info().SizeClass >= core.SizeClassLarge {
					return ui.HStack(
						views...,
					).FullWidth().
						Gap(ui.L16).
						Alignment(ui.Stretch)
				}
				return ui.VStack(
					views...,
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

func (c TCalendar) mapStartTimeCluster(vp ViewPort, events []Event) []cluster {
	switch vp.style {
	case vpYear:
		var tmp []cluster
		for _, week := range mapStartTimeClusterCalWeeks(c.vp, c.events) {
			tmp = append(tmp, week)
		}

		return tmp
	default:
		var tmp []cluster
		for _, time := range mapStartTimeClusterCalTime(vp, events) {
			tmp = append(tmp, time)
		}

		return tmp
	}

}

type seqPillTimeHint int

const (
	pillTimeYYYYMMDD seqPillTimeHint = iota
	pillTimeHHMM
	pillTimeNone
)

func startTimeSeqPill(c TCalendar, evt Event, timeHint seqPillTimeHint, isLarge bool) core.View {
	colors := c.colors

	isTimePoint := evt.To.At.Equal(evt.From.At)

	var timeStr string
	if timeHint != pillTimeNone {

		if isTimePoint {
			if timeHint == pillTimeHHMM {
				timeStr = evt.From.At.Format("15:04")
			} else {
				if evt.To.At.Hour() == 0 && evt.To.At.Minute() == 0 && evt.From.At.Hour() == 0 && evt.To.At.Minute() == 0 {
					timeStr = evt.From.At.Format(xtime.GermanDate)
				} else {
					timeStr = evt.From.At.Format(xtime.GermanDateTime)
				}
			}
		} else {
			if timeHint == pillTimeHHMM {
				timeStr = evt.From.At.Format("15:04") + " - " + evt.To.At.Format("15:04")
			} else {
				if evt.To.At.Hour() == 0 && evt.To.At.Minute() == 0 && evt.From.At.Hour() == 0 && evt.To.At.Minute() == 0 {
					timeStr = evt.From.At.Format(xtime.GermanDate) + " - " + evt.To.At.Format(xtime.GermanDate)
				} else {
					timeStr = evt.From.At.Format(xtime.GermanDateTime) + " - " + evt.To.At.Format(xtime.GermanDateTime)
				}
			}
		}
	}

	titleRow := []core.View{ui.Text(evt.Label).Font(ui.BodyLarge)}
	if isLarge && len(evt.Chips) > 0 {
		titleRow = append(titleRow,
			ui.Spacer(),
			ui.HStack(
				chipViews(evt)...,
			).
				Alignment(ui.Trailing).
				Gap(ui.L12),
		)
	}

	return ui.HStack(
		// category color
		ui.HStack().BackgroundColor(evt.Category.Color).Frame(ui.Frame{MinWidth: ui.L12}).AccessibilityLabel(evt.Category.Label),
		ui.VStack(
			ui.HStack(
				titleRow...,
			).Alignment(ui.Leading).FullWidth(),
			ui.If(timeStr != "",
				ui.HStack(
					ui.ImageIcon(icons.Clock),
					ui.Text(timeStr),
				).Gap(ui.L4),
			),
			ui.If(evt.Organiser != "", ui.HStack(
				ui.ImageIcon(icons.User),
				ui.Text(evt.Organiser),
			).Gap(ui.L4)),
			ui.If(evt.Location != "", ui.HStack(
				ui.ImageIcon(icons.MapPinAlt),
				ui.Text(evt.Location),
			).Gap(ui.L4)),
			ui.Lazy(func() core.View {
				if isLarge {
					if evt.AttendeeState != nil {
						return ui.HStack(
							chipView(*evt.AttendeeState),
						).Alignment(ui.Leading).FullWidth()
					}
					return nil
				}
				if evt.AttendeeState != nil {
					return ui.HStack(
						chipView(*evt.AttendeeState),
					).Alignment(ui.Leading).FullWidth()
				}
				return ui.HStack(
					chipViews(evt)...,
				).Alignment(ui.Leading).Gap(ui.L12)
			}),
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
		FullWidth().
		Alignment(ui.Stretch).
		Border(ui.Border{}.Radius(ui.L8))
}

func chipViews(evt Event) []core.View {
	return ui.ForEach[Chip, core.View](evt.Chips, func(chip Chip) core.View {
		return chipView(chip)
	})
}

func chipView(chip Chip) core.View {
	return ui.HStack(
		ui.ImageIcon(chip.Icon).FillColor(chip.TextColor).StrokeColor(chip.TextColor).Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
		ui.Text(chip.Label).Font(ui.Font{Size: ui.L12}).Color(chip.TextColor),
	).
		Alignment(ui.Leading).
		BackgroundColor(chip.BgColor).
		Gap(ui.L4).
		FullWidth().
		Border(ui.Border{}.Radius(ui.L4)).
		Padding(ui.Padding{}.All(ui.L4))
}
