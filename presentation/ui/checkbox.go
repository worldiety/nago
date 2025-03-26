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

type TCheckbox struct {
	value      bool
	inputValue *core.State[bool]
	disabled   bool
	invisible  bool
	id         string
}

// Checkbox represents a user interface element which spans a visible area to click or tap from the user.
// Use it for controls, which do not cause an immediate effect. See also [Toggle].
func Checkbox(checked bool) TCheckbox {
	c := TCheckbox{
		value: checked,
	}

	return c
}

func (c TCheckbox) InputChecked(input *core.State[bool]) TCheckbox {
	c.inputValue = input
	return c
}

func (c TCheckbox) Disabled(disabled bool) TCheckbox {
	c.disabled = disabled
	return c
}

func (c TCheckbox) Visible(v bool) TCheckbox {
	c.invisible = !v
	return c
}

func (c TCheckbox) ID(id string) TCheckbox {
	c.id = id
	return c
}

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
