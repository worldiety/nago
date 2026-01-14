// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// TCheckbox is a basic component (Checkbox).
// It allows users to toggle between checked and unchecked states,
// optionally binding to external state. The checkbox can be disabled,
// hidden, or assigned a unique identifier for reference.
type TCheckbox struct {
	value      bool              // current checked state (true = checked, false = unchecked)
	inputValue *core.State[bool] // optional external binding for controlled state
	disabled   bool              // when true, interaction is disabled
	invisible  bool              // when true, the checkbox is not rendered
	id         string            // unique identifier for the checkbox
}

// Checkbox represents a user interface element which spans a visible area to click or tap from the user.
// Use it for controls, which do not cause an immediate effect. See also [Toggle].
func Checkbox(checked bool) TCheckbox {
	c := TCheckbox{
		value: checked,
	}

	return c
}

// Deprecated: use InputValue
// InputChecked binds the checkbox to an external boolean state,
// allowing it to be controlled from outside the component.
func (c TCheckbox) InputChecked(input *core.State[bool]) TCheckbox {
	c.inputValue = input
	return c
}

// InputValue binds the checkbox to an external boolean state,
// allowing it to be controlled from outside the component.
func (c TCheckbox) InputValue(input *core.State[bool]) TCheckbox {
	c.inputValue = input
	return c
}

// Disabled enables or disables user interaction with the checkbox.
func (c TCheckbox) Disabled(disabled bool) TCheckbox {
	c.disabled = disabled
	return c
}

// Visible controls the visibility of the checkbox; setting false hides it.
func (c TCheckbox) Visible(v bool) TCheckbox {
	c.invisible = !v
	return c
}

// ID assigns a unique identifier to the checkbox, useful for testing or referencing.
func (c TCheckbox) ID(id string) TCheckbox {
	c.id = id
	return c
}

// Render builds and returns the protocol representation of the checkbox.
func (c TCheckbox) Render(ctx core.RenderContext) core.RenderNode {
	// TODO this component has an intrinsic padding which must be removed
	return &proto.Checkbox{
		Value:      proto.Bool(c.value),
		InputValue: c.inputValue.Ptr(),
		Disabled:   proto.Bool(c.disabled),
		Invisible:  proto.Bool(c.invisible),
		Id:         proto.Str(c.id),
	}
}
