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

// TRichTextEditor is an editor for rich text, see RichTextEditor. It supports two-way data binding, and layout
// configuration by a frame.
type TRichTextEditor struct{ ref js.Ref }

func (TRichTextEditor) isView() {}

// RichTextEditor returns a rich text editor with the initial value.
//
//gostrict:host
func RichTextEditor(value string) TRichTextEditor {
	return TRichTextEditor{ref: wrap(nui.RichTextEditor(value))}
}

// InputValue binds the content of the editor to the state, which keeps the text between two renders.
//
//gostrict:host
func (c TRichTextEditor) InputValue(state StringState) TRichTextEditor {
	return TRichTextEditor{ref: wrap(get[nui.TRichTextEditor](c.ref).InputValue(state.nago()))}
}

// Frame sets the layout frame of the editor, e.g. its size.
//
//gostrict:host
func (c TRichTextEditor) Frame(frame Frame) TRichTextEditor {
	return TRichTextEditor{ref: wrap(get[nui.TRichTextEditor](c.ref).Frame(frame.nago()))}
}

// FullWidth expands the editor to the full available width.
//
//gostrict:host
func (c TRichTextEditor) FullWidth() TRichTextEditor {
	return TRichTextEditor{ref: wrap(get[nui.TRichTextEditor](c.ref).FullWidth())}
}
