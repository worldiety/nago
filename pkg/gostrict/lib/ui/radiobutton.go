// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// RadioStateGroup is a group of boolean states of the window, of which exactly one is selected, see
// AutoRadioStateGroup and RadioButtonField.
type RadioStateGroup struct{ ref js.Ref }

// AutoRadioStateGroup creates a group of boolean states for radio-button-like behavior. Exactly one state at a time
// is selected; when a state becomes selected, all others are cleared. The states are stored under window-scoped
// keys derived from the id and the index. A negative number of states is none.
//
//gostrict:host
func AutoRadioStateGroup(wnd Window, id string, states int32) RadioStateGroup {
	return b2GroupFrom(nui.AutoRadioStateGroup(wnd.nago(), id, int(max(states, 0))))
}

// InitIndex selects the state at idx, if the states are new.
//
//gostrict:host
func (s RadioStateGroup) InitIndex(idx int32) RadioStateGroup {
	return b2GroupFrom(b2GroupOf(s).InitIndex(int(idx)))
}

// Observe calls f with the index of the selected state, whenever any state of the group changes.
//
//gostrict:host
func (s RadioStateGroup) Observe(f func(newIdx int32)) {
	b2GroupOf(s).Observe(func(newIdx int) { f(int32(newIdx)) })
}

// SetSelectedIndex selects the state at idx and clears all others.
//
//gostrict:host
func (s RadioStateGroup) SetSelectedIndex(idx int32) {
	b2GroupOf(s).SetSelectedIndex(int(idx))
}

// Notify triggers the observers of all states of the group.
//
//gostrict:host
func (s RadioStateGroup) Notify() {
	b2GroupOf(s).Notify()
}

// SelectedIndex returns the index of the selected state or -1, if no state is selected.
//
//gostrict:host
func (s RadioStateGroup) SelectedIndex() int32 {
	return int32(b2GroupOf(s).SelectedIndex())
}

// TRadioButton is a radio button, see RadioButton.
type TRadioButton struct{ ref js.Ref }

func (TRadioButton) isView() {}

// RadioButton represents a user interface element, which the user clicks or taps to select exactly one of several
// options. Use it for controls, which do not cause an immediate effect. See also Toggle and Checkbox.
//
//gostrict:host
func RadioButton(checked bool) TRadioButton {
	return TRadioButton{ref: wrap(nui.RadioButton(checked))}
}

// Value sets the initial selected state of the radio button.
//
//gostrict:host
func (c TRadioButton) Value(checked bool) TRadioButton {
	return TRadioButton{ref: wrap(get[nui.TRadioButton](c.ref).Value(checked))}
}

// InputChecked binds the radio button to a state of the window, which keeps the selected state in sync.
//
//gostrict:host
func (c TRadioButton) InputChecked(input BoolState) TRadioButton {
	return TRadioButton{ref: wrap(get[nui.TRadioButton](c.ref).InputChecked(input.nago()))}
}

// ID sets the id of the radio button, e.g. for testing or referencing.
//
//gostrict:host
func (c TRadioButton) ID(id string) TRadioButton {
	return TRadioButton{ref: wrap(get[nui.TRadioButton](c.ref).ID(id))}
}

// Disabled disables the radio button, which prevents the user interaction.
//
//gostrict:host
func (c TRadioButton) Disabled(disabled bool) TRadioButton {
	return TRadioButton{ref: wrap(get[nui.TRadioButton](c.ref).Disabled(disabled))}
}

// Visible shows or hides the radio button.
//
//gostrict:host
func (c TRadioButton) Visible(visible bool) TRadioButton {
	return TRadioButton{ref: wrap(get[nui.TRadioButton](c.ref).Visible(visible))}
}

// Name sets the name of the radio button.
//
//gostrict:host
func (c TRadioButton) Name(name string) TRadioButton {
	return TRadioButton{ref: wrap(get[nui.TRadioButton](c.ref).Name(name))}
}
