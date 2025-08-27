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

// TRichTextEditor is a composite component (Rich Text Editor).
// It provides an interactive editor for creating and modifying rich text content.
// The editor supports two-way data binding, read-only and disabled states,
// and layout configuration via frame settings.
type TRichTextEditor struct {
	value      string              // initial text content
	inputValue *core.State[string] // optional bound state for two-way updates
	frame      Frame               // layout frame
	readOnly   bool                // makes the editor read-only if true
	disabled   bool                // disables interaction if true
}

// RichTextEditor creates a new rich text editor with the given initial value.
func RichTextEditor(value string) TRichTextEditor {
	return TRichTextEditor{
		value: value,
	}
}

// InputValue binds the editor's content to a state, enabling two-way data binding.
func (c TRichTextEditor) InputValue(state *core.State[string]) TRichTextEditor {
	c.inputValue = state
	return c
}

// Frame sets the layout frame of the editor.
func (c TRichTextEditor) Frame(frame Frame) TRichTextEditor {
	c.frame = frame
	return c
}

// FullWidth expands the editor to take the full available width.
func (c TRichTextEditor) FullWidth() TRichTextEditor {
	c.frame = c.frame.FullWidth()
	return c
}

// Render builds and returns the protocol representation of the rich text editor,
// including its content, state binding, read-only/disabled flags, and frame.
func (c TRichTextEditor) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.RichTextEditor{
		Value:      proto.Str(c.value),
		Frame:      c.frame.ora(),
		ReadOnly:   proto.Bool(c.readOnly),
		Disabled:   proto.Bool(c.disabled),
		InputValue: c.inputValue.Ptr(),
	}
}
