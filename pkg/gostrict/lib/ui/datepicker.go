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
)

// TDatePicker lets the user select a day or a range of days, like TDatePicker of nago.
type TDatePicker struct{ ref js.Ref }

func (TDatePicker) isView() {}

// SingleDatePicker returns a picker for a single day, whose value the state holds, like SingleDatePicker of nago.
//
//gostrict:host
func SingleDatePicker(label string, value Date, state DateState) TDatePicker {
	return TDatePicker{ref: wrap(nui.SingleDatePicker(label, value.nago(), state.nago()))}
}

// RangeDatePicker returns a picker for a range of days, whose first and last day the states hold, like
// RangeDatePicker of nago.
//
//gostrict:host
func RangeDatePicker(label string, startValue Date, startState DateState, endValue Date, endState DateState) TDatePicker {
	return TDatePicker{ref: wrap(nui.RangeDatePicker(label, startValue.nago(), startState.nago(), endValue.nago(), endState.nago()))}
}

// Padding sets the padding.
//
//gostrict:host
func (c TDatePicker) Padding(padding Padding) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).Padding(padding.nago()).(nui.TDatePicker))}
}

// Frame sets the size.
//
//gostrict:host
func (c TDatePicker) Frame(frame Frame) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).Frame(frame.nago()).(nui.TDatePicker))}
}

// WithFrame changes the size with fn, which gets the current one.
//
//gostrict:host
func (c TDatePicker) WithFrame(fn func(Frame) Frame) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TDatePicker))}
}

// Border sets the border.
//
//gostrict:host
func (c TDatePicker) Border(border Border) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).Border(border.nago()).(nui.TDatePicker))}
}

// Visible shows or hides the view.
//
//gostrict:host
func (c TDatePicker) Visible(visible bool) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).Visible(visible).(nui.TDatePicker))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TDatePicker) AccessibilityLabel(label string) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).AccessibilityLabel(label).(nui.TDatePicker))}
}

// Disabled enables or disables the interaction.
//
//gostrict:host
func (c TDatePicker) Disabled(disabled bool) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).Disabled(disabled))}
}

// SupportingText sets the helper text below the picker.
//
//gostrict:host
func (c TDatePicker) SupportingText(text string) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).SupportingText(text))}
}

// ErrorText sets the validation message.
//
//gostrict:host
func (c TDatePicker) ErrorText(text string) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).ErrorText(text))}
}

// DoubleMode shows two months instead of one.
//
//gostrict:host
func (c TDatePicker) DoubleMode(doubleMode bool) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).DoubleMode(doubleMode))}
}

// Optional marks the field as optional.
//
//gostrict:host
func (c TDatePicker) Optional(optional bool) TDatePicker {
	return TDatePicker{ref: wrap(get[nui.TDatePicker](c.ref).Optional(optional))}
}
