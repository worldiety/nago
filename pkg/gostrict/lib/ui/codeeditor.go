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

// TCodeEditor is a code editor, see CodeEditor. It is optimized for writing and displaying code, e.g. with syntax
// highlighting and a tab size of 4 spaces.
type TCodeEditor struct{ ref js.Ref }

func (TCodeEditor) isView() {}

// CodeEditor returns a code editor with the initial value and a tab size of 4 spaces.
//
//gostrict:host
func CodeEditor(value string) TCodeEditor {
	return TCodeEditor{ref: wrap(nui.CodeEditor(value))}
}

// Value sets the initial text content of the code editor.
//
//gostrict:host
func (c TCodeEditor) Value(value string) TCodeEditor {
	return TCodeEditor{ref: wrap(get[nui.TCodeEditor](c.ref).Value(value))}
}

// InputValue binds the editor to the state, which keeps the text between two renders.
//
//gostrict:host
func (c TCodeEditor) InputValue(state StringState) TCodeEditor {
	return TCodeEditor{ref: wrap(get[nui.TCodeEditor](c.ref).InputValue(state.nago()))}
}

// Disabled enables or disables the user interaction with the editor.
//
//gostrict:host
func (c TCodeEditor) Disabled(b bool) TCodeEditor {
	return TCodeEditor{ref: wrap(get[nui.TCodeEditor](c.ref).Disabled(b))}
}

// Frame sets the layout frame of the editor, including its size.
//
//gostrict:host
func (c TCodeEditor) Frame(frame Frame) TCodeEditor {
	return TCodeEditor{ref: wrap(get[nui.TCodeEditor](c.ref).Frame(frame.nago()))}
}

// FullWidth expands the editor to the full available width.
//
//gostrict:host
func (c TCodeEditor) FullWidth() TCodeEditor {
	return TCodeEditor{ref: wrap(get[nui.TCodeEditor](c.ref).FullWidth())}
}

// Language gives a hint for the syntax highlighting, e.g. go, html, css, json, xml or markdown. Other languages may be
// supported as well.
//
//gostrict:host
func (c TCodeEditor) Language(language string) TCodeEditor {
	return TCodeEditor{ref: wrap(get[nui.TCodeEditor](c.ref).Language(language))}
}
