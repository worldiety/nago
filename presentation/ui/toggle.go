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

// TToggle is a basic component (Toggle).
// This component represents a switch-like control (on/off) without a label.
// It is intended for immediate activation or deactivation of features.
type TToggle struct {
	id         string
	value      bool              // current toggle state (true = on, false = off)
	inputValue *core.State[bool] // optional bound state for two-way binding
	disabled   bool              // disables interaction when true
	invisible  bool              // hides the toggle when true
}

// Toggle is just a kind of checkbox without a label. However, a toggle shall be used for immediate activation
// functions. In contrast to that, use a checkbox for form things without an immediate effect.
func Toggle(checked bool) TToggle {
	c := TToggle{
		value: checked,
	}

	return c
}

// ID sets the ID of the toggle
func (c TToggle) ID(id string) TToggle {
	c.id = id
	return c
}

// InputChecked binds the toggle to an external boolean state for two-way data binding.
func (c TToggle) InputChecked(input *core.State[bool]) TToggle {
	c.inputValue = input
	return c
}

// Disabled enables or disables interaction with the toggle.
func (c TToggle) Disabled(disabled bool) TToggle {
	c.disabled = disabled
	return c
}

// Visible controls the visibility of the toggle; false hides it from the UI.
func (c TToggle) Visible(v bool) TToggle {
	c.invisible = !v
	return c
}

// Render builds and returns the protocol node for the toggle, including current value,
// optional bound state, disabled and visibility flags.
func (c TToggle) Render(ctx core.RenderContext) core.RenderNode {
	// TODO toggle has a screwed intrinsic padding/offset into top
	return &proto.Toggle{
		InputValue: c.inputValue.Ptr(),
		Value:      proto.Bool(c.value),
		Disabled:   proto.Bool(c.disabled),
		Invisible:  proto.Bool(c.invisible),
		Id:         proto.Str(c.id),
	}
}
