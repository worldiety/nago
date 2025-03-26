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

type TToggle struct {
	value      bool
	inputValue *core.State[bool]
	disabled   bool
	invisible  bool
}

// Toggle is just a kind of checkbox without a label. However, a toggle shall be used for immediate activation
// functions. In contrast to that, use a checkbox for form things without an immediate effect.
func Toggle(checked bool) TToggle {
	c := TToggle{
		value: checked,
	}

	return c
}

func (c TToggle) InputChecked(input *core.State[bool]) TToggle {
	c.inputValue = input
	return c
}

func (c TToggle) Disabled(disabled bool) TToggle {
	c.disabled = disabled
	return c
}

func (c TToggle) Visible(v bool) TToggle {
	c.invisible = !v
	return c
}

func (c TToggle) Render(ctx core.RenderContext) core.RenderNode {
	// TODO toggle has a screwed intrinsic padding/offset into top
	return &proto.Toggle{
		Value:      proto.Bool(c.value),
		InputValue: c.inputValue.Ptr(),
		Disabled:   proto.Bool(c.disabled),
		Invisible:  proto.Bool(c.invisible),
	}
}
