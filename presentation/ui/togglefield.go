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

type TToggleField struct {
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

// A ToggleField aggregates a toggle together with form field typical labels, hints and error texts.
func ToggleField(label string, value bool) TToggleField {
	return TToggleField{
		label: label,
		value: value,
	}
}

func (c TToggleField) Padding(padding Padding) DecoredView {
	c.padding = padding
	return c
}

func (c TToggleField) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

func (c TToggleField) Border(border Border) DecoredView {
	c.border = border
	return c
}

func (c TToggleField) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TToggleField) AccessibilityLabel(label string) DecoredView {
	c.label = label
	return c
}

func (c TToggleField) InputValue(inputValue *core.State[bool]) TToggleField {
	c.inputValue = inputValue
	return c
}

func (c TToggleField) SupportingText(text string) TToggleField {
	c.supportingText = text
	return c
}

func (c TToggleField) ErrorText(text string) TToggleField {
	c.errorText = text
	return c
}

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
