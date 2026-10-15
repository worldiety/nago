// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timeframe"
)

// TimeframePickerFormat is the display format of a timeframe picker, like timeframe.PickerFormat of nago.
type TimeframePickerFormat int32

const (
	// TimeframeClassicFormat selects a date, a start time and an end time and shows the duration, like
	// timeframe.ClassicFormat of nago.
	TimeframeClassicFormat TimeframePickerFormat = 0
)

// TTimeframePicker lets the user select a date with a start and an end time, like timeframe.TPicker of nago.
type TTimeframePicker struct{ ref js.Ref }

func (TTimeframePicker) isView() {}

// TimeframePicker returns a picker with the label for the time frame of the state, like timeframe.Picker of nago. An
// empty state is set to the current time.
//
//gostrict:host
func TimeframePicker(label string, state TimeFrameState) TTimeframePicker {
	return TTimeframePicker{ref: wrap(timeframe.Picker(label, state.nago()))}
}

// Padding sets the padding.
//
//gostrict:host
func (c TTimeframePicker) Padding(padding Padding) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).Padding(padding.nago()).(timeframe.TPicker))}
}

// Frame sets the size.
//
//gostrict:host
func (c TTimeframePicker) Frame(frame Frame) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).Frame(frame.nago()).(timeframe.TPicker))}
}

// WithFrame changes the size with fn, which gets the current one.
//
//gostrict:host
func (c TTimeframePicker) WithFrame(fn func(Frame) Frame) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(timeframe.TPicker))}
}

// Border sets the border.
//
//gostrict:host
func (c TTimeframePicker) Border(border Border) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).Border(border.nago()).(timeframe.TPicker))}
}

// Visible shows or hides the view.
//
//gostrict:host
func (c TTimeframePicker) Visible(visible bool) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).Visible(visible).(timeframe.TPicker))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TTimeframePicker) AccessibilityLabel(label string) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).AccessibilityLabel(label).(timeframe.TPicker))}
}

// Disabled enables or disables the interaction.
//
//gostrict:host
func (c TTimeframePicker) Disabled(disabled bool) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).Disabled(disabled))}
}

// Title sets the title of the dialog.
//
//gostrict:host
func (c TTimeframePicker) Title(title string) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).Title(title))}
}

// SupportingText sets the helper text below the picker.
//
//gostrict:host
func (c TTimeframePicker) SupportingText(text string) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).SupportingText(text))}
}

// ErrorText sets the validation message.
//
//gostrict:host
func (c TTimeframePicker) ErrorText(text string) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).ErrorText(text))}
}

// Format sets the display format.
//
//gostrict:host
func (c TTimeframePicker) Format(format TimeframePickerFormat) TTimeframePicker {
	return TTimeframePicker{ref: wrap(get[timeframe.TPicker](c.ref).Format(timeframe.PickerFormat(format)))}
}
