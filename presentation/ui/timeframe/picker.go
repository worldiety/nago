// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeframe

import (
	"math"
	"time"

	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timepicker"
)

type PickerFormat int

const (
	// ClassicFormat allows to select a date, a start time and end time and displays the selected duration.
	ClassicFormat PickerFormat = iota
)

// TPicker is a util component (Time Frame Picker).
// It allows users to pick a date and a start/end time, optionally binding
// the result to an external state. The picker supports different formats,
// validation messages, and can be configured with a specific time zone.
type TPicker struct {
	label          string                       // primary label shown with the control
	supportingText string                       // helper or secondary text shown below the label
	errorText      string                       // validation or error message
	frame          ui.Frame                     // layout frame for size and positioning
	day            *core.State[xtime.Date]      // selected date
	startTime      *core.State[time.Duration]   // start time of the frame
	endTime        *core.State[time.Duration]   // end time of the frame
	targetState    *core.State[xtime.TimeFrame] // external binding for the full time frame
	title          string                       // title used when presented in a dialog
	format         PickerFormat                 // defines display/interaction format for the picker
	disabled       bool                         // when true, interaction is disabled
	tz             *time.Location               // time zone used for interpreting times
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

// roundToMinute rounds the given Unix timestamp in milliseconds
// to the nearest whole minute.
func roundToMinute(t xtime.UnixMilliseconds) xtime.UnixMilliseconds {
	return xtime.UnixMilliseconds(math.Round(float64(t)/1000/60)) * 1000 * 60
}

// Padding sets the inner spacing around the picker content.
// (currently not implemented)
func (c TPicker) Padding(padding ui.Padding) ui.DecoredView {
	//TODO implement me
	return c
}

// Frame sets the layout frame of the picker, including size and positioning.
func (c TPicker) Frame(frame ui.Frame) ui.DecoredView {
	c.frame = frame
	return c
}

// WithFrame applies a transformation function to the picker's frame
// and returns the updated component.
func (c TPicker) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	c.frame = fn(c.frame)
	return c
}

// Border sets the border style of the picker.
// (currently not implemented)
func (c TPicker) Border(border ui.Border) ui.DecoredView {
	//TODO implement me
	return c
}

// Visible controls the visibility of the picker; setting false hides it.
// (currently not implemented)
func (c TPicker) Visible(visible bool) ui.DecoredView {
	//TODO implement me
	return c
}

// AccessibilityLabel sets a label used by screen readers for accessibility.
// (currently not implemented)
func (c TPicker) AccessibilityLabel(label string) ui.DecoredView {
	//TODO implement me
	return c
}

// Disabled enables or disables user interaction with the picker.
func (c TPicker) Disabled(disabled bool) TPicker {
	c.disabled = disabled
	return c
}

// Title sets the title of the picker, typically shown in dialogs.
func (c TPicker) Title(title string) TPicker {
	c.title = title
	return c
}

// Format sets the picker format, which controls its display and interaction style.
func (c TPicker) Format(format PickerFormat) TPicker {
	c.format = format
	return c
}

// SupportingText sets helper or secondary text displayed below the picker label.
func (c TPicker) SupportingText(text string) TPicker {
	c.supportingText = text
	return c
}

// ErrorText sets the validation or error message displayed below the picker.
func (c TPicker) ErrorText(text string) TPicker {
	c.errorText = text
	return c
}

// Render builds and returns the UI representation of the time frame picker.
// It displays a date picker, start and end time pickers, and a read-only field
// showing the calculated duration. The component also renders labels,
// supporting text, or error messages depending on its state.
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
