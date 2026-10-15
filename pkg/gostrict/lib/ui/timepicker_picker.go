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
	"go.wdy.de/nago/presentation/ui/timepicker"
)

// TTimepickerPicker lets the user choose a duration with the granularity of days, hours, minutes and seconds, like
// timepicker.TPicker of nago.
type TTimepickerPicker struct{ ref js.Ref }

func (TTimepickerPicker) isView() {}

// TimepickerPicker returns a picker with the label for the duration of the state, like timepicker.Picker of nago. It
// shows hours and minutes by default.
//
//gostrict:host
func TimepickerPicker(label string, state DurationState) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(timepicker.Picker(label, state.nago()))}
}

// Padding sets the padding.
//
//gostrict:host
func (c TTimepickerPicker) Padding(padding Padding) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Padding(padding.nago()).(timepicker.TPicker))}
}

// Frame sets the size.
//
//gostrict:host
func (c TTimepickerPicker) Frame(frame Frame) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Frame(frame.nago()).(timepicker.TPicker))}
}

// WithFrame changes the size with fn, which gets the current one.
//
//gostrict:host
func (c TTimepickerPicker) WithFrame(fn func(Frame) Frame) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(timepicker.TPicker))}
}

// Border sets the border.
//
//gostrict:host
func (c TTimepickerPicker) Border(border Border) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Border(border.nago()).(timepicker.TPicker))}
}

// Visible shows or hides the view.
//
//gostrict:host
func (c TTimepickerPicker) Visible(visible bool) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Visible(visible).(timepicker.TPicker))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TTimepickerPicker) AccessibilityLabel(label string) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).AccessibilityLabel(label).(timepicker.TPicker))}
}

// Disabled enables or disables the interaction.
//
//gostrict:host
func (c TTimepickerPicker) Disabled(disabled bool) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Disabled(disabled))}
}

// Title sets the title of the dialog.
//
//gostrict:host
func (c TTimepickerPicker) Title(title string) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Title(title))}
}

// SupportingText sets the helper text below the picker.
//
//gostrict:host
func (c TTimepickerPicker) SupportingText(text string) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).SupportingText(text))}
}

// ErrorText sets the validation message.
//
//gostrict:host
func (c TTimepickerPicker) ErrorText(text string) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).ErrorText(text))}
}

// Hours shows or hides the selection of hours.
//
//gostrict:host
func (c TTimepickerPicker) Hours(showHours bool) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Hours(showHours))}
}

// Minutes shows or hides the selection of minutes.
//
//gostrict:host
func (c TTimepickerPicker) Minutes(showMinutes bool) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Minutes(showMinutes))}
}

// Days shows or hides the selection of days.
//
//gostrict:host
func (c TTimepickerPicker) Days(showDays bool) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Days(showDays))}
}

// Seconds shows or hides the selection of seconds.
//
//gostrict:host
func (c TTimepickerPicker) Seconds(showSeconds bool) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Seconds(showSeconds))}
}

// Format sets the display format.
//
//gostrict:host
func (c TTimepickerPicker) Format(format TimepickerPickerFormat) TTimepickerPicker {
	return TTimepickerPicker{ref: wrap(get[timepicker.TPicker](c.ref).Format(timepicker.PickerFormat(format)))}
}
