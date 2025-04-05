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

type TRichTextEditor struct {
	value      string
	inputValue *core.State[string]
	frame      Frame

	readOnly bool
	disabled bool
}

func RichTextEditor(value string) TRichTextEditor {
	return TRichTextEditor{
		value: value,
	}
}

func (c TRichTextEditor) InputValue(state *core.State[string]) TRichTextEditor {
	c.inputValue = state
	return c
}

func (c TRichTextEditor) Frame(frame Frame) TRichTextEditor {
	c.frame = frame
	return c
}

func (c TRichTextEditor) FullWidth() TRichTextEditor {
	c.frame = c.frame.FullWidth()
	return c
}

func (c TRichTextEditor) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.RichTextEditor{
		Value:      proto.Str(c.value),
		Frame:      c.frame.ora(),
		ReadOnly:   proto.Bool(c.readOnly),
		Disabled:   proto.Bool(c.disabled),
		InputValue: c.inputValue.Ptr(),
	}
}
