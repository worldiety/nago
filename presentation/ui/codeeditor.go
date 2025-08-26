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

// TCodeEditor is a composite component(Code Editor).
// This component provides a text editor interface
// optimized for writing and displaying code. It supports syntax highlighting,
// configurable tab size, and optional read-only or disabled states.
type TCodeEditor struct {
	value      string
	inputValue *core.State[string]
	frame      Frame
	language   string
	readOnly   bool
	disabled   bool
	tabSize    int
}

// CodeEditor creates a new code editor with the given initial value
// and a default tab size of 4 spaces.
func CodeEditor(value string) TCodeEditor {
	return TCodeEditor{
		value:   value,
		tabSize: 4,
	}
}

// InputValue binds the editor to an external state for controlled text value updates.
func (c TCodeEditor) InputValue(state *core.State[string]) TCodeEditor {
	c.inputValue = state
	return c
}

// Disabled enables or disables user interaction with the editor.
func (c TCodeEditor) Disabled(b bool) TCodeEditor {
	c.disabled = b
	return c
}

// Frame sets the layout frame of the editor, including size and positioning.
func (c TCodeEditor) Frame(frame Frame) TCodeEditor {
	c.frame = frame
	return c
}

// FullWidth sets the editor to span the full available width.
func (c TCodeEditor) FullWidth() TCodeEditor {
	c.frame = c.frame.FullWidth()
	return c
}

// Language gives a syntax highlighting hint.
// Defined are go, html, css, json, xml, markdown but there may be arbitrary support.
func (c TCodeEditor) Language(language string) TCodeEditor {
	c.language = language
	return c
}

// Render builds and returns the protocol representation of the code editor.
func (c TCodeEditor) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.CodeEditor{
		Value:      proto.Str(c.value),
		Frame:      c.frame.ora(),
		ReadOnly:   proto.Bool(c.readOnly),
		Disabled:   proto.Bool(c.disabled),
		TabSize:    proto.Uint(c.tabSize),
		InputValue: c.inputValue.Ptr(),
		Language:   proto.Str(c.language),
	}
}
