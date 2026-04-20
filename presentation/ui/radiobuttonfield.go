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

// TRadioButtonField is a basic component (RadioButton Field).
// It combines a radio button with a label
// The field can be bound to external state and visibility controls.
type TRadioButtonField struct {
	value      bool              // whether the radio button is selected
	inputValue *core.State[bool] // optional bound state for two-way data binding
	disabled   bool              // disables interaction when true
	invisible  bool              // hides the radio button when true
	id         string
	label      string
	stateGroup *RadioStateGroup
	index      int
	name       string
}

// RadioButtonField combines a RadioButton with a label
func RadioButtonField(label string, stateGroup *RadioStateGroup, index int) TRadioButtonField {
	var inputValue *core.State[bool]
	for i, state := range stateGroup.All() {
		if i == index {
			inputValue = state
		}
	}

	c := TRadioButtonField{
		value:      stateGroup.SelectedIndex() == index,
		inputValue: inputValue,
		label:      label,
		stateGroup: stateGroup,
		index:      index,
	}

	return c
}

// Label sets the label of the radio button field
func (c TRadioButtonField) Label(label string) TRadioButtonField {
	c.label = label
	return c
}

// InputChecked binds the radio button to the given state, enabling two-way data binding
// so that the selected state is synchronized with external logic.
func (c TRadioButtonField) InputChecked(input *core.State[bool]) TRadioButtonField {
	c.inputValue = input
	return c
}

func (c TRadioButtonField) ID(id string) TRadioButtonField {
	c.id = id
	return c
}

// Disabled disables the radio button when set to true, preventing user interaction.
func (c TRadioButtonField) Disabled(disabled bool) TRadioButtonField {
	c.disabled = disabled
	return c
}

// Visible controls the visibility of the radio button.
// Passing false will hide the component from the UI.
func (c TRadioButtonField) Visible(v bool) TRadioButtonField {
	c.invisible = !v
	return c
}

// Name assigns a name to the checkbox field, useful for autocomplete
func (c TRadioButtonField) Name(name string) TRadioButtonField {
	c.name = name
	return c
}

// Render builds and returns the protocol representation of the radio button.
func (c TRadioButtonField) Render(ctx core.RenderContext) core.RenderNode {
	action := func() {
		c.stateGroup.SetSelectedIndex(c.index)
	}
	opacity := 1.0
	if c.disabled {
		action = nil
		opacity = 0.6
	}

	return HStack(
		RadioButton(c.value).
			InputChecked(c.inputValue).Disabled(c.disabled).ID(c.id).Visible(!c.invisible).Name(c.name),
		HStack(
			Text(c.label).Action(action),
		).Opacity(opacity),
	).Gap(L4).Render(ctx)
}
