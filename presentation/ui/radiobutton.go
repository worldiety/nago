// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"
	"iter"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type RadioStateGroup struct {
	states []*core.State[bool]
}

// AutoRadioStateGroup creates a group of boolean states for radio-button-like behavior.
// Exactly one state at a time is set to true; when any state changes to true,
// all others are set to false. The states are stored under window-scoped keys
// derived from the provided id and the index.
func AutoRadioStateGroup(wnd core.Window, id string, states int) RadioStateGroup {
	var wndStates = make([]*core.State[bool], 0, states)
	for i := range states {
		state := core.StateOf[bool](wnd, fmt.Sprintf("%s-%d", id, i))
		state.Observe(func(newValue bool) {
			for i2, wndState := range wndStates {
				wndState.Set(i == i2)
			}
		})
		wndStates = append(wndStates, state)
	}

	return RadioStateGroup{states: wndStates}
}

// Observe registers a callback that is invoked with the currently selected index
// whenever any state in the group changes.
func (s RadioStateGroup) Observe(f func(newIdx int)) {
	for _, state := range s.states {
		state.Observe(func(newValue bool) {
			f(s.SelectedIndex())
		})
	}
}

// SetSelectedIndex marks the state at idx as selected (true) and clears all others.
func (s RadioStateGroup) SetSelectedIndex(idx int) {
	for i, state := range s.states {
		state.Set(idx == i)
	}
}

// Notify triggers observers of all states in the group.
func (s RadioStateGroup) Notify() {
	for _, state := range s.states {
		state.Notify()
	}
}

// SelectedIndex returns -1 or the selected index.
func (s RadioStateGroup) SelectedIndex() int {
	for i, state := range s.states {
		if state.Get() {
			return i
		}
	}

	return -1
}

// All iterates over all states in the group, yielding (index, state) pairs.
func (s RadioStateGroup) All() iter.Seq2[int, *core.State[bool]] {
	return func(yield func(int, *core.State[bool]) bool) {
		for i := range s.states {
			if !yield(i, s.states[i]) {
				return
			}
		}
	}
}

// TRadioButton is a basic component (Radio Button).
// It represents a selectable option in a group where only one element can be active at a time.
// Radio buttons are typically used in forms or settings where the user must pick exactly one choice.
type TRadioButton struct {
	value      bool              // whether the radio button is selected
	inputValue *core.State[bool] // optional bound state for two-way data binding
	disabled   bool              // disables interaction when true
	invisible  bool              // hides the radio button when true
	id         string
}

// RadioButton represents a user interface element which spans a visible area to click or tap from the user.
// Use it for controls, which do not cause an immediate effect and only one element can be picked at a time.
// See also [Toggle], [Checkbox] and [Select].
func RadioButton(checked bool) TRadioButton {
	c := TRadioButton{
		value: checked,
	}

	return c
}

// InputChecked binds the radio button to the given state, enabling two-way data binding
// so that the selected state is synchronized with external logic.
func (c TRadioButton) InputChecked(input *core.State[bool]) TRadioButton {
	c.inputValue = input
	return c
}

func (c TRadioButton) ID(id string) TRadioButton {
	c.id = id
	return c
}

// Disabled disables the radio button when set to true, preventing user interaction.
func (c TRadioButton) Disabled(disabled bool) TRadioButton {
	c.disabled = disabled
	return c
}

// Visible controls the visibility of the radio button.
// Passing false will hide the component from the UI.
func (c TRadioButton) Visible(v bool) TRadioButton {
	c.invisible = !v
	return c
}

// Render builds and returns the protocol representation of the radio button.
func (c TRadioButton) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.Radiobutton{
		Value:      proto.Bool(c.value),
		InputValue: c.inputValue.Ptr(),
		Disabled:   proto.Bool(c.disabled),
		Invisible:  proto.Bool(c.invisible),
		Id:         proto.Str(c.id),
	}
}
