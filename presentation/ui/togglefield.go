// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"
	"math/rand"

	"go.wdy.de/nago/presentation/core"
)

// TToggleField is a composite component (Toggle Field).
// This component combines a toggle with form-related elements such as
// a label, supporting text, and error messages.
type TToggleField struct {
	id              string
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

// ID sets the ID of the toggle field
func (c TToggleField) ID(id string) TToggleField {
	c.id = id
	return c
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

// Disabled sets the disabled state of the toggle field
func (c TToggleField) Disabled(disabled bool) TToggleField {
	c.disabled = disabled
	return c
}

// Render builds and returns the visual representation of the toggle field.
func (c TToggleField) Render(context core.RenderContext) core.RenderNode {
	if c.id == "" && c.inputValue != nil {
		c.id = fmt.Sprintf("cb%d", rand.Intn(999999999999)) // random id fallback
	}

	labelFor := c.id
	opacity := 1.0
	if c.disabled {
		labelFor = ""
		opacity = 0.6
	}

	return HStack(
		Toggle(c.value).
			ID(c.id).
			Disabled(c.disabled).
			InputChecked(c.inputValue),
		VStack(
			Text(c.label).LabelFor(labelFor),
			IfElse(c.errorText == "",
				Text(c.supportingText).Font(Font{Size: "0.75rem"}).Color(ST0).LabelFor(labelFor),
				Text(c.errorText).Font(Font{Size: "0.75rem"}).Color(SE0).LabelFor(labelFor),
			),
		).Alignment(Leading).Opacity(opacity),
	).Alignment(Leading).
		Gap(L8).
		Border(c.border).
		Visible(!c.invisible).
		Padding(c.padding).
		Frame(c.frame).Render(context)
}
