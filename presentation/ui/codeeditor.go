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

type TCodeEditor struct {
	value      string
	inputValue *core.State[string]
	frame      Frame
	language   string
	readOnly   bool
	disabled   bool
	tabSize    int
}

func CodeEditor(value string) TCodeEditor {
	return TCodeEditor{
		value:   value,
		tabSize: 4,
	}
}

func (c TCodeEditor) InputValue(state *core.State[string]) TCodeEditor {
	c.inputValue = state
	return c
}

func (c TCodeEditor) Disabled(b bool) TCodeEditor {
	c.disabled = b
	return c
}

func (c TCodeEditor) Frame(frame Frame) TCodeEditor {
	c.frame = frame
	return c
}

func (c TCodeEditor) FullWidth() TCodeEditor {
	c.frame = c.frame.FullWidth()
	return c
}

// Language gives a syntax highlighting hint. Defined are go, html, css, json, xml,
// markdown but there may be arbitrary support.
func (c TCodeEditor) Language(language string) TCodeEditor {
	c.language = language
	return c
}

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
