package ui

import (
	"go.wdy.de/nago/presentation/core"
)

type TCheckboxField struct {
	label           string
	value           bool
	inputValue      *core.State[bool]
	supportingText  string
	errorText       string
	disabled        bool
	invisible       bool
	padding         Padding
	frame           Frame
	border          Border
	keyboardOptions TKeyboardOptions
}

// A CheckboxField aggregates a checkbox together with form field typical labels, hints and error texts.
func CheckboxField(label string, value bool) TCheckboxField {
	return TCheckboxField{
		label: label,
		value: value,
	}
}

func (c TCheckboxField) Padding(padding Padding) DecoredView {
	c.padding = padding
	return c
}

func (c TCheckboxField) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

func (c TCheckboxField) Border(border Border) DecoredView {
	c.border = border
	return c
}

func (c TCheckboxField) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TCheckboxField) AccessibilityLabel(label string) DecoredView {
	c.label = label
	return c
}

func (c TCheckboxField) InputValue(inputValue *core.State[bool]) TCheckboxField {
	c.inputValue = inputValue
	return c
}

func (c TCheckboxField) SupportingText(text string) TCheckboxField {
	c.supportingText = text
	return c
}

func (c TCheckboxField) ErrorText(text string) TCheckboxField {
	c.errorText = text
	return c
}

func (c TCheckboxField) Render(context core.RenderContext) core.RenderNode {
	return VStack(
		HStack(
			Checkbox(c.value).
				Disabled(c.disabled).
				InputChecked(c.inputValue),
			Text(c.label),
		),
		IfElse(c.errorText == "",
			Text(c.supportingText).Font(Font{Size: "0.75rem"}).Color(ST0),
			Text(c.errorText).Font(Font{Size: "0.75rem"}).Color(SE0),
		),
	).Alignment(Leading).
		Border(c.border).
		Visible(!c.invisible).
		Padding(c.padding).
		Frame(c.frame).Render(context)
}
