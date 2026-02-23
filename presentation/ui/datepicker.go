// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// TDatePicker is a composite component (Date Picker).
// It allows users to select either a single date or a date range,
// depending on its style. The component supports external state
// bindings, validation messages, and layout configuration.
type TDatePicker struct {
	label                   string                  // label displayed above the date picker
	disabled                bool                    // when true, interaction is disabled
	invisible               bool                    // when true, the picker is not rendered
	style                   proto.DatePickerStyle   // defines picker type (single date or range)
	supportingText          string                  // helper or secondary text shown below the picker
	errorText               string                  // validation or error message
	startOrSingleValue      xtime.Date              // selected start date or single date
	inputStartOrSingleValue *core.State[xtime.Date] // external binding for start/single date
	endValue                xtime.Date              // selected end date (only for range)
	inputEndValue           *core.State[xtime.Date] // external binding for end date (only for range)
	frame                   Frame                   // layout frame for sizing and positioning
	doubleMode              bool                    // when true, the picker shows two months instead of one
}

// SingleDatePicker creates a date picker configured for selecting a single date,
// binding the given value and optional state.
func SingleDatePicker(label string, value xtime.Date, inputValue *core.State[xtime.Date]) TDatePicker {
	return TDatePicker{
		label:                   label,
		style:                   proto.DatePickerSingleDate,
		inputStartOrSingleValue: inputValue,
		startOrSingleValue:      value,
	}
}

// RangeDatePicker creates a date picker configured for selecting a date range,
// binding start and end values to their respective states.
func RangeDatePicker(label string, startValue xtime.Date, startInputValue *core.State[xtime.Date], endValue xtime.Date, endInputValue *core.State[xtime.Date]) TDatePicker {
	return TDatePicker{
		label:                   label,
		style:                   proto.DatePickerDateRange,
		inputStartOrSingleValue: startInputValue,
		startOrSingleValue:      startValue,
		endValue:                endValue,
		inputEndValue:           endInputValue,
	}
}

// Padding sets the inner spacing around the date picker content.
func (c TDatePicker) Padding(padding Padding) DecoredView {
	//TODO implement me
	return c
}

// Frame sets the layout frame of the date picker, including size and positioning.
func (c TDatePicker) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

// WithFrame applies a transformation function to the picker's frame
// and returns the updated component.
func (c TDatePicker) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// Border sets the border styling of the date picker.
func (c TDatePicker) Border(border Border) DecoredView {
	//TODO implement me
	return c
}

// Visible controls the visibility of the date picker; setting false hides it.
func (c TDatePicker) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// AccessibilityLabel sets a label used by screen readers for accessibility.
func (c TDatePicker) AccessibilityLabel(label string) DecoredView {
	//TODO implement me
	return c
}

// Disabled enables or disables user interaction with the date picker.
func (c TDatePicker) Disabled(disabled bool) TDatePicker {
	c.disabled = disabled
	return c
}

// SupportingText sets helper or secondary text displayed below the picker label.
func (c TDatePicker) SupportingText(text string) TDatePicker {
	c.supportingText = text
	return c
}

// ErrorText sets the validation or error message displayed below the picker.
func (c TDatePicker) ErrorText(text string) TDatePicker {
	c.errorText = text
	return c
}

// DoubleMode enables double-month mode for range pickers.
func (c TDatePicker) DoubleMode(doubleMode bool) TDatePicker {
	c.doubleMode = doubleMode
	return c
}

// Render builds and returns the protocol representation of the date picker.
func (c TDatePicker) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.DatePicker{
		Disabled:       proto.Bool(c.disabled),
		Label:          proto.Str(c.label),
		SupportingText: proto.Str(c.supportingText),
		ErrorText:      proto.Str(c.errorText),
		Style:          c.style,
		Value: proto.DateData{
			Day:   proto.Day(c.startOrSingleValue.Day),
			Month: proto.Month(c.startOrSingleValue.Month),
			Year:  proto.Year(c.startOrSingleValue.Year),
		},
		InputValue: c.inputStartOrSingleValue.Ptr(),
		EndValue: proto.DateData{
			Day:   proto.Day(c.endValue.Day),
			Month: proto.Month(c.endValue.Month),
			Year:  proto.Year(c.endValue.Year),
		},
		Frame:         c.frame.ora(),
		EndInputValue: c.inputEndValue.Ptr(),
		Invisible:     proto.Bool(c.invisible),
		DoubleMode:    proto.Bool(c.doubleMode),
	}
}
