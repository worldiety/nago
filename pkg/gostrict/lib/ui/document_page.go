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
	"go.wdy.de/nago/presentation/ui/document"
)

// DocumentSize is the size of a page, like document.Size of nago.
type DocumentSize struct {
	Width  Length
	Height Length
}

// DocumentDinA4 is the size of a DIN A4 page, like document.DinA4 of nago.
//
//gostrict:js ({ Width: "50rem", Height: "70rem" })
var DocumentDinA4 DocumentSize = documentSizeFrom(document.DinA4)

// TDocumentPage is a page of an endless paper document, like document.TPage of nago. See DocumentPage.
type TDocumentPage struct{ ref js.Ref }

func (TDocumentPage) isView() {}

// DocumentPage creates a virtual endless DIN A4 styled paper document with the items, like document.Page of nago.
//
//gostrict:host
func DocumentPage(items ...View) TDocumentPage {
	return TDocumentPage{ref: wrap(document.Page(viewsOf(items)...))}
}

// Size sets the page size to fixed width and height constraints.
//
//gostrict:host
func (c TDocumentPage) Size(size DocumentSize) TDocumentPage {
	return TDocumentPage{ref: wrap(get[document.TPage](c.ref).Size(size.nago()))}
}

// Alignment sets how the children of the page are aligned.
//
//gostrict:host
func (c TDocumentPage) Alignment(alignment Alignment) TDocumentPage {
	return TDocumentPage{ref: wrap(get[document.TPage](c.ref).Alignment(nui.Alignment(alignment)))}
}

// Border applies a border to the page.
//
//gostrict:host
func (c TDocumentPage) Border(border Border) TDocumentPage {
	return TDocumentPage{ref: wrap(get[document.TPage](c.ref).Border(border.nago()))}
}

// Frame directly sets the frame of the page.
//
//gostrict:host
func (c TDocumentPage) Frame(frame Frame) TDocumentPage {
	return TDocumentPage{ref: wrap(get[document.TPage](c.ref).Frame(frame.nago()))}
}

// BackgroundColor sets the background color of the page.
//
//gostrict:host
func (c TDocumentPage) BackgroundColor(color Color) TDocumentPage {
	return TDocumentPage{ref: wrap(get[document.TPage](c.ref).BackgroundColor(nui.Color(color)))}
}

// Comment appends the comments to the existing comments of the page.
//
//gostrict:host
func (c TDocumentPage) Comment(comments ...TDocumentComment) TDocumentPage {
	return TDocumentPage{ref: wrap(get[document.TPage](c.ref).Comment(documentCommentsOf(comments)...))}
}

// Append appends the views to the items of the page.
//
//gostrict:host
func (c TDocumentPage) Append(items ...View) TDocumentPage {
	return TDocumentPage{ref: wrap(get[document.TPage](c.ref).Append(viewsOf(items)...))}
}
