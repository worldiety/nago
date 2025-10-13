// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
)

// TCheckboxField is a basic component (Checkbox Field).
// It combines a checkbox with a label, supporting text, and optional
// error messages. The field can be bound to external state and styled
// with padding, frame, and border. It also supports accessibility,
// keyboard options, and visibility controls.
type TCheckboxField struct {
	label              string
	value              bool
	inputValue         *core.State[bool]
	supportingText     string
	errorText          string
	disabled           bool
	invisible          bool
	padding            Padding
	frame              Frame
	border             Border
	keyboardOptions    TKeyboardOptions
	accessibilityLabel string
	id                 string
}

// A CheckboxField aggregates a checkbox together with form field typical labels, hints and error texts.
func CheckboxField(label string, value bool) TCheckboxField {
	return TCheckboxField{
		label: label,
		value: value,
	}
}

// Enabled sets whether the checkbox field is interactive.
// Equivalent to Disabled(!b).
func (c TCheckboxField) Enabled(b bool) TCheckboxField {
	c.disabled = !b
	return c
}

// Disabled enables or disables user interaction with the checkbox field.
func (c TCheckboxField) Disabled(b bool) TCheckboxField {
	c.disabled = b
	return c
}

// Padding sets the inner spacing around the checkbox field.
func (c TCheckboxField) Padding(padding Padding) DecoredView {
	c.padding = padding
	return c
}

// Frame sets the layout frame of the checkbox field, including size and positioning.
func (c TCheckboxField) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

// WithFrame applies a transformation function to the field's frame
// and returns the updated component.
func (c TCheckboxField) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// Border sets the border styling of the checkbox field.
func (c TCheckboxField) Border(border Border) DecoredView {
	c.border = border
	return c
}

// Visible controls the visibility of the checkbox field; setting false hides it.
func (c TCheckboxField) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// AccessibilityLabel sets the label used for accessibility purposes.
func (c TCheckboxField) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// InputValue binds the checkbox field to an external boolean state.
func (c TCheckboxField) InputValue(inputValue *core.State[bool]) TCheckboxField {
	c.inputValue = inputValue
	return c
}

// SupportingText sets helper or secondary text shown below the label.
func (c TCheckboxField) SupportingText(text string) TCheckboxField {
	c.supportingText = text
	return c
}

// ErrorText sets the validation or error message displayed below the field.
func (c TCheckboxField) ErrorText(text string) TCheckboxField {
	c.errorText = text
	return c
}

// ID assigns a unique identifier to the checkbox field, useful for testing or referencing.
func (c TCheckboxField) ID(id string) TCheckboxField {
	c.id = id
	return c
}

// Render builds and returns the UI representation of the checkbox field.
func (c TCheckboxField) Render(context core.RenderContext) core.RenderNode {
	return VStack(
		HStack(
			Checkbox(c.value).
				ID(c.id).
				Disabled(c.disabled).
				InputChecked(c.inputValue),
			Text(c.label),
		),
		IfElse(c.errorText == "",
			Text(c.supportingText).Font(Font{Size: "0.75rem"}).Color(ST0),
			Text(c.errorText).Font(Font{Size: "0.75rem"}).Color(SE0),
		),
	).Alignment(Leading).
		AccessibilityLabel(c.accessibilityLabel).
		Border(c.border).
		Visible(!c.invisible).
		Padding(c.padding).
		Frame(c.frame).Render(context)
}
