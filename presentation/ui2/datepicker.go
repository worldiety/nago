package ui

import (
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
	startOrSingleValue      ora.Date
	inputStartOrSingleValue *core.State[ora.Date]
	endValue                ora.Date
	inputEndValue           *core.State[ora.Date]
}

func SingleDatePicker(label string, value ora.Date, inputValue *core.State[ora.Date]) TDatePicker {
	return TDatePicker{
		label:                   label,
		style:                   ora.DatePickerSingleDate,
		inputStartOrSingleValue: inputValue,
		startOrSingleValue:      value,
	}
}

func RangeDatePicker(label string, startValue ora.Date, startInputValue *core.State[ora.Date], endValue ora.Date, endInputValue *core.State[ora.Date]) TDatePicker {
	return TDatePicker{
		label:                   label,
		style:                   ora.DatePickerDateRange,
		inputStartOrSingleValue: startInputValue,
		startOrSingleValue:      startValue,
		endValue:                endValue,
		inputEndValue:           endInputValue,
	}
}

func (c TDatePicker) Render(ctx core.RenderContext) ora.Component {
	return ora.DatePicker{
		Type:           ora.DatePickerT,
		Disabled:       false,
		Label:          c.label,
		SupportingText: c.supportingText,
		ErrorText:      c.errorText,
		Style:          c.style,
		Value:          c.startOrSingleValue,
		InputValue:     c.inputStartOrSingleValue.Ptr(),
		EndValue:       c.endValue,
		EndInputValue:  c.inputEndValue.Ptr(),
		Invisible:      c.invisible,
	}
}
