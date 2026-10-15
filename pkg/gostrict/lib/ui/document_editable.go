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
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/document"
)

// DocumentToggleStyle is the style of the toggle button of a DocumentEditable, like document.ToggleStyle of nago.
type DocumentToggleStyle int32

const (
	// DocumentTopTrailing places the toggle button above the content, aligned to the top right, like document.TopTrailing of nago.
	DocumentTopTrailing DocumentToggleStyle = 0
	// DocumentInlineTrailing places the toggle button inline to the right of the content, like document.InlineTrailing of nago.
	DocumentInlineTrailing DocumentToggleStyle = 1
	// DocumentClickable makes the whole component clickable to enter the edit mode, like document.Clickable of nago.
	DocumentClickable DocumentToggleStyle = 2
)

// TDocumentEditable toggles between a view mode and an edit mode, like document.TEditable of nago. See DocumentEditable.
type TDocumentEditable struct{ ref js.Ref }

func (TDocumentEditable) isView() {}

// DocumentEditable creates an editable component with the view and the edit callbacks, like document.Editable of nago.
//
//gostrict:host
func DocumentEditable(onView func() View, onEdit func() View) TDocumentEditable {
	return TDocumentEditable{ref: wrap(document.Editable(
		func() core.View { return viewOf(onView()) },
		func() core.View { return viewOf(onEdit()) },
	))}
}

// InputValue binds the state, which tells whether the edit mode is presented.
//
//gostrict:host
func (c TDocumentEditable) InputValue(editPresented BoolState) TDocumentEditable {
	return TDocumentEditable{ref: wrap(get[document.TEditable](c.ref).InputValue(editPresented.nago()))}
}

// Style sets the toggle style.
//
//gostrict:host
func (c TDocumentEditable) Style(style DocumentToggleStyle) TDocumentEditable {
	return TDocumentEditable{ref: wrap(get[document.TEditable](c.ref).Style(document.ToggleStyle(int(style))))}
}

// Comment sets the action, which is executed when a comment is triggered.
//
//gostrict:host
func (c TDocumentEditable) Comment(action func()) TDocumentEditable {
	return TDocumentEditable{ref: wrap(get[document.TEditable](c.ref).Comment(action))}
}

// Frame sets the frame of the editable component.
//
//gostrict:host
func (c TDocumentEditable) Frame(frame Frame) TDocumentEditable {
	return TDocumentEditable{ref: wrap(get[document.TEditable](c.ref).Frame(frame.nago()))}
}

// BackgroundColor sets the background color of the editable component.
//
//gostrict:host
func (c TDocumentEditable) BackgroundColor(color Color) TDocumentEditable {
	return TDocumentEditable{ref: wrap(get[document.TEditable](c.ref).BackgroundColor(nui.Color(color)))}
}

// Alignment sets the alignment of the editable component.
//
//gostrict:host
func (c TDocumentEditable) Alignment(alignment Alignment) TDocumentEditable {
	return TDocumentEditable{ref: wrap(get[document.TEditable](c.ref).Alignment(nui.Alignment(alignment)))}
}

// Border sets the border of the editable component.
//
//gostrict:host
func (c TDocumentEditable) Border(border Border) TDocumentEditable {
	return TDocumentEditable{ref: wrap(get[document.TEditable](c.ref).Border(border.nago()))}
}

// FullWidth expands the editable component to the full width.
//
//gostrict:host
func (c TDocumentEditable) FullWidth() TDocumentEditable {
	return TDocumentEditable{ref: wrap(get[document.TEditable](c.ref).FullWidth())}
}
