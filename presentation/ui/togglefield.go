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

// TToggleField is a composite component (Toggle Field).
// This component combines a toggle with form-related elements such as
// a label, supporting text, and error messages.
type TToggleField struct {
	label           string            // label displayed next to the toggle
	value           bool              // initial toggle state (true = on, false = off)
	inputValue      *core.State[bool] // optional bound state for two-way binding
	supportingText  string            // optional hint or helper text
	errorText       string            // optional validation error message
	disabled        bool              // disables interaction when true
	invisible       bool              // hides the field when true
	padding         Padding           // custom inner padding
	frame           Frame             // layout frame
	border          Border            // border styling
	keyboardOptions TKeyboardOptions  // keyboard configuration
}

// A ToggleField aggregates a toggle together with form field typical labels, hints and error texts.
func ToggleField(label string, value bool) TToggleField {
	return TToggleField{
		label: label,
		value: value,
	}
}

// Padding sets the inner padding of the toggle field.
func (c TToggleField) Padding(padding Padding) DecoredView {
	c.padding = padding
	return c
}

// Frame sets the layout frame of the toggle field.
func (c TToggleField) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

// WithFrame modifies the layout frame using the provided function.
func (c TToggleField) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// Border sets the border styling of the toggle field.
func (c TToggleField) Border(border Border) DecoredView {
	c.border = border
	return c
}

// Visible controls the visibility of the toggle field.
func (c TToggleField) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// AccessibilityLabel sets the accessibility label for screen readers.
func (c TToggleField) AccessibilityLabel(label string) DecoredView {
	c.label = label
	return c
}

// InputValue binds the toggle field to a reactive state for two-way binding.
func (c TToggleField) InputValue(inputValue *core.State[bool]) TToggleField {
	c.inputValue = inputValue
	return c
}

// SupportingText sets optional supporting text displayed below the field.
func (c TToggleField) SupportingText(text string) TToggleField {
	c.supportingText = text
	return c
}

// ErrorText sets the error message displayed when validation fails.
func (c TToggleField) ErrorText(text string) TToggleField {
	c.errorText = text
	return c
}

// Render builds and returns the visual representation of the toggle field.
func (c TToggleField) Render(context core.RenderContext) core.RenderNode {
	return VStack(
		HStack(
			Toggle(c.value).
				Disabled(c.disabled).
				InputChecked(c.inputValue),
			Text(c.label).
				Padding(Padding{Bottom: L8}), // TODO remove we, as soon as toggle is fixed
		).Gap(L8).
			Padding(Padding{}.All(L8)),

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
