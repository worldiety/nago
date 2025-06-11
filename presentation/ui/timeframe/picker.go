// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeframe

import (
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timepicker"
	"math"
	"time"
)

type PickerFormat int

const (
	// ClassicFormat allows to select a date, a start time and end time and displays the selected duration.
	ClassicFormat PickerFormat = iota
)

type TPicker struct {
	label          string
	supportingText string
	errorText      string
	frame          ui.Frame
	day            *core.State[xtime.Date]
	startTime      *core.State[time.Duration]
	endTime        *core.State[time.Duration]
	targetState    *core.State[xtime.TimeFrame]
	title          string
	format         PickerFormat
	disabled       bool
	tz             *time.Location
}

// Picker renders a xtime.TimeFrame picker to select at least a day and a start and end time (inclusive).
func Picker(label string, selectedState *core.State[xtime.TimeFrame]) TPicker {
	if selectedState.Get().IsZero() {
		selectedState.Set(xtime.TimeFrame{
			StartTime: xtime.Now(),
			EndTime:   xtime.Now(),
		})
	}

	tz := selectedState.Get().Timezone.Location()

	day := selectedState.Get().StartTime.Date(tz)
	startOffset := selectedState.Get().StartTime.Time(tz).Sub(day.Time(tz)).Truncate(time.Minute)
	endOffset := selectedState.Get().EndTime.Time(tz).Sub(day.Time(tz)).Truncate(time.Minute)

	p := TPicker{
		label:       label,
		format:      ClassicFormat,
		targetState: selectedState,
		tz:          tz,
		day: core.DerivedState[xtime.Date](selectedState, "day").Init(func() xtime.Date {
			return day
		}),
		startTime: core.DerivedState[time.Duration](selectedState, "start").Init(func() time.Duration {
			return startOffset
		}),
		endTime: core.DerivedState[time.Duration](selectedState, "end").Init(func() time.Duration {
			return endOffset
		}),
	}

	p.day.Observe(func(newValue xtime.Date) {
		p.startTime.Notify()
		p.endTime.Notify()
	})

	p.startTime.Observe(func(newValue time.Duration) {
		if p.startTime.Get() > p.endTime.Get() {
			p.endTime.Set(p.endTime.Get() + 24*time.Hour)
			p.endTime.Notify()
		}

		day := p.day.Get().Time(p.tz)

		tf := p.targetState.Get()
		tf.StartTime = roundToMinute(xtime.UnixMilliseconds(newValue.Milliseconds() + day.UnixMilli()))
		tf.EndTime = roundToMinute(xtime.UnixMilliseconds(p.endTime.Get().Milliseconds() + day.UnixMilli()))
		tf.Timezone = xtime.Timezone(tz.String())
		p.targetState.Set(tf)
		p.targetState.Notify()
	})

	p.endTime.Observe(func(newValue time.Duration) {
		if p.startTime.Get() > p.endTime.Get() {
			p.endTime.Set(p.endTime.Get() + 24*time.Hour)
			p.endTime.Notify()
		}

		day := p.day.Get().Time(p.tz)

		tf := p.targetState.Get()
		tf.StartTime = roundToMinute(xtime.UnixMilliseconds(p.startTime.Get().Milliseconds() + day.UnixMilli()))
		tf.EndTime = roundToMinute(xtime.UnixMilliseconds(newValue.Milliseconds() + day.UnixMilli()))
		tf.Timezone = xtime.Timezone(tz.String())
		p.targetState.Set(tf)
		p.targetState.Notify()
	})

	return p
}

func roundToMinute(t xtime.UnixMilliseconds) xtime.UnixMilliseconds {
	return xtime.UnixMilliseconds(math.Round(float64(t)/1000/60)) * 1000 * 60
}

func (c TPicker) Padding(padding ui.Padding) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker) Frame(frame ui.Frame) ui.DecoredView {
	c.frame = frame
	return c
}

func (c TPicker) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	c.frame = fn(c.frame)
	return c
}

func (c TPicker) Border(border ui.Border) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker) Visible(visible bool) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker) AccessibilityLabel(label string) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker) Disabled(disabled bool) TPicker {
	c.disabled = disabled
	return c
}

func (c TPicker) Title(title string) TPicker {
	c.title = title
	return c
}

func (c TPicker) Format(format PickerFormat) TPicker {
	c.format = format
	return c
}

func (c TPicker) SupportingText(text string) TPicker {
	c.supportingText = text
	return c
}

func (c TPicker) ErrorText(text string) TPicker {
	c.errorText = text
	return c
}

func (c TPicker) Render(ctx core.RenderContext) core.RenderNode {
	var duration string
	if c.targetState.Get().Empty() {
		duration = "kein Zeitraum gewählt"
	} else {
		duration = (c.endTime.Get() - c.startTime.Get()).String()
	}

	inner := ui.VStack(
		ui.SingleDatePicker("", c.day.Get(), c.day).Frame(ui.Frame{}.FullWidth()),
		ui.Grid(
			ui.GridCell(timepicker.Picker("Startzeit wählen", c.startTime).Hours(true).Minutes(true)),
			ui.GridCell(timepicker.Picker("Endzeit wählen", c.endTime).Hours(true).Minutes(true)),
			ui.GridCell(ui.TextField("Dauer", duration).Disabled(true)),
		).Gap(ui.L8).
			Rows(1).
			FullWidth(),
	).Gap(ui.L8).FullWidth()

	return ui.VStack(
		ui.IfElse(c.errorText == "",
			ui.Text(c.label).Font(ui.Font{Size: ui.L16}),
			ui.HStack(
				ui.Image().StrokeColor(ui.SE0).Embed(heroSolid.XMark).Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
				ui.Text("").Font(ui.Font{Size: ui.L16}).Color(ui.SE0),
			),
		),
		inner,
		ui.IfElse(c.errorText == "",
			ui.Text(c.supportingText).Font(ui.Font{Size: "0.75rem"}).Color(ui.ST0),
			ui.Text(c.errorText).Font(ui.Font{Size: "0.75rem"}).Color(ui.SE0),
		),
	).Alignment(ui.Leading).
		Gap(ui.L4).
		Frame(c.frame).
		Render(ctx)
}
