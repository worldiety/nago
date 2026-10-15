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
	"go.wdy.de/nago/presentation/ui/list"
)

// TListEntry is a row of a list, like list.TEntry of nago. See ListEntry.
type TListEntry struct{ ref js.Ref }

func (TListEntry) isView() {}

// ListEntry creates a full width entry, like list.Entry of nago.
//
//gostrict:host
func ListEntry() TListEntry {
	return TListEntry{ref: wrap(list.Entry())}
}

// Headline sets the main title text of the entry.
//
//gostrict:host
func (c TListEntry) Headline(s string) TListEntry {
	return TListEntry{ref: wrap(get[list.TEntry](c.ref).Headline(s))}
}

// SupportingText sets an optional supporting text below the headline.
//
//gostrict:host
func (c TListEntry) SupportingText(s string) TListEntry {
	return TListEntry{ref: wrap(get[list.TEntry](c.ref).SupportingText(s))}
}

// SupportingView sets an optional supporting view below the headline.
//
//gostrict:host
func (c TListEntry) SupportingView(view View) TListEntry {
	return TListEntry{ref: wrap(get[list.TEntry](c.ref).SupportingView(viewOf(view)))}
}

// HeadlineView sets an optional view, which replaces the headline text.
//
//gostrict:host
func (c TListEntry) HeadlineView(view View) TListEntry {
	return TListEntry{ref: wrap(get[list.TEntry](c.ref).HeadlineView(viewOf(view)))}
}

// Leading sets an optional leading view, e.g. an icon or an avatar.
//
//gostrict:host
func (c TListEntry) Leading(v View) TListEntry {
	return TListEntry{ref: wrap(get[list.TEntry](c.ref).Leading(viewOf(v)))}
}

// Trailing sets an optional trailing view, e.g. a button or a chevron.
//
//gostrict:host
func (c TListEntry) Trailing(v View) TListEntry {
	return TListEntry{ref: wrap(get[list.TEntry](c.ref).Trailing(viewOf(v)))}
}

// Action sets the action, which is called when the user clicks the entry.
//
//gostrict:host
func (c TListEntry) Action(fn func()) TListEntry {
	return TListEntry{ref: wrap(get[list.TEntry](c.ref).Action(fn))}
}

// Frame sets the layout frame of the entry.
//
//gostrict:host
func (c TListEntry) Frame(frame Frame) TListEntry {
	return TListEntry{ref: wrap(get[list.TEntry](c.ref).Frame(frame.nago()))}
}

// TList is a vertical collection of rows, like list.TList of nago. See List.
type TList struct{ ref js.Ref }

func (TList) isView() {}

// List creates a list with the entries as rows, like list.List of nago.
//
//gostrict:host
func List(entries ...View) TList {
	return TList{ref: wrap(list.List(viewsOf(entries)...))}
}

// ColorBody sets the color of the body.
//
//gostrict:host
func (c TList) ColorBody(color Color) TList {
	return TList{ref: wrap(get[list.TList](c.ref).ColorBody(nui.Color(color)))}
}

// ColorCaption sets the color of the caption.
//
//gostrict:host
func (c TList) ColorCaption(color Color) TList {
	return TList{ref: wrap(get[list.TList](c.ref).ColorCaption(nui.Color(color)))}
}

// ColorHighlight sets the color of the highlighted rows.
//
//gostrict:host
func (c TList) ColorHighlight(color Color) TList {
	return TList{ref: wrap(get[list.TList](c.ref).ColorHighlight(nui.Color(color)))}
}

// ColorHover sets the color of the rows, while the pointer is over them.
//
//gostrict:host
func (c TList) ColorHover(color Color) TList {
	return TList{ref: wrap(get[list.TList](c.ref).ColorHover(nui.Color(color)))}
}

// ColorFooter sets the color of the footer.
//
//gostrict:host
func (c TList) ColorFooter(color Color) TList {
	return TList{ref: wrap(get[list.TList](c.ref).ColorFooter(nui.Color(color)))}
}

// Caption sets an optional caption view above the list.
//
//gostrict:host
func (c TList) Caption(s View) TList {
	return TList{ref: wrap(get[list.TList](c.ref).Caption(viewOf(s)))}
}

// Frame sets the layout frame of the list.
//
//gostrict:host
func (c TList) Frame(frame Frame) TList {
	return TList{ref: wrap(get[list.TList](c.ref).Frame(frame.nago()))}
}

// FullWidth expands the list to the full available width.
//
//gostrict:host
func (c TList) FullWidth() TList {
	return TList{ref: wrap(get[list.TList](c.ref).FullWidth())}
}

// Footer sets an optional footer view below the list.
//
//gostrict:host
func (c TList) Footer(s View) TList {
	return TList{ref: wrap(get[list.TList](c.ref).Footer(viewOf(s)))}
}

// OnEntryClicked sets the callback, which is called with the index of the row, which the user clicks.
//
//gostrict:host
func (c TList) OnEntryClicked(fn func(idx int32)) TList {
	return TList{ref: wrap(get[list.TList](c.ref).OnEntryClicked(func(idx int) { fn(int32(idx)) }))}
}

// OnHighlighted sets the callback, which tells for the index of a row, whether the row is highlighted.
//
//gostrict:host
func (c TList) OnHighlighted(fn func(idx int32) bool) TList {
	return TList{ref: wrap(get[list.TList](c.ref).OnHighlighted(func(idx int) bool { return fn(int32(idx)) }))}
}

// With changes the list with fn.
//
//gostrict:host
func (c TList) With(fn func(l TList) TList) TList {
	return TList{ref: wrap(get[list.TList](c.ref).With(func(v list.TList) list.TList {
		return get[list.TList](fn(TList{ref: wrap(v)}).ref)
	}))}
}
