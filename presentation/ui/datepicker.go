package ui

import (
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TDatePicker struct {
	label                   string
	disabled                bool
	invisible               bool
	style                   ora.DatePickerStyle
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
		style:                   ora.DatePickerSingleDate,
		inputStartOrSingleValue: inputValue,
		startOrSingleValue:      value,
	}
}

func RangeDatePicker(label string, startValue xtime.Date, startInputValue *core.State[xtime.Date], endValue xtime.Date, endInputValue *core.State[xtime.Date]) TDatePicker {
	return TDatePicker{
		label:                   label,
		style:                   ora.DatePickerDateRange,
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

func (c TDatePicker) Render(ctx core.RenderContext) ora.Component {
	return ora.DatePicker{
		Type:           ora.DatePickerT,
		Disabled:       c.disabled,
		Label:          c.label,
		SupportingText: c.supportingText,
		ErrorText:      c.errorText,
		Style:          c.style,
		Value: ora.Date{
			Day:   c.startOrSingleValue.Day,
			Month: int(c.startOrSingleValue.Month),
			Year:  c.startOrSingleValue.Year,
		},
		InputValue: c.inputStartOrSingleValue.Ptr(),
		EndValue: ora.Date{
			Day:   c.endValue.Day,
			Month: int(c.endValue.Month),
			Year:  c.endValue.Year,
		},
		Frame:         c.frame.ora(),
		EndInputValue: c.inputEndValue.Ptr(),
		Invisible:     c.invisible,
	}
}
