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

type TDatePicker struct {
	label                   string
	disabled                bool
	invisible               bool
	style                   proto.DatePickerStyle
	supportingText          string
	errorText               string
	startOrSingleValue      xtime.Date
	inputStartOrSingleValue *core.State[xtime.Date]
	endValue                xtime.Date
	inputEndValue           *core.State[xtime.Date]
	frame                   Frame
}

func SingleDatePicker(label string, value xtime.Date, inputValue *core.State[xtime.Date]) TDatePicker {
	return TDatePicker{
		label:                   label,
		style:                   proto.DatePickerSingleDate,
		inputStartOrSingleValue: inputValue,
		startOrSingleValue:      value,
	}
}

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

func (c TDatePicker) Padding(padding Padding) DecoredView {
	//TODO implement me
	return c
}

func (c TDatePicker) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

func (c TDatePicker) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

func (c TDatePicker) Border(border Border) DecoredView {
	//TODO implement me
	return c
}

func (c TDatePicker) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TDatePicker) AccessibilityLabel(label string) DecoredView {
	//TODO implement me
	return c
}

func (c TDatePicker) Disabled(disabled bool) TDatePicker {
	c.disabled = disabled
	return c
}

func (c TDatePicker) SupportingText(text string) TDatePicker {
	c.supportingText = text
	return c
}

func (c TDatePicker) ErrorText(text string) TDatePicker {
	c.errorText = text
	return c
}

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
	}
}
