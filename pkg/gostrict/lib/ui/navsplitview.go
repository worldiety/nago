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
	"go.wdy.de/nago/presentation/ui/navsplitview"
)

// NavsplitviewViewID identifies a view of a split view, like navsplitview.ViewID of nago.
type NavsplitviewViewID string

// NavsplitviewTargetKind is the kind of a column, which shows a view of a split view, like navsplitview.TargetKind
// of nago.
type NavsplitviewTargetKind int32

const (
	// NavsplitviewKindSidebar is the column of the sidebar.
	NavsplitviewKindSidebar NavsplitviewTargetKind = 1
	// NavsplitviewKindContent is the column of the content.
	NavsplitviewKindContent NavsplitviewTargetKind = 2
	// NavsplitviewKindDetail is the column of the detail.
	NavsplitviewKindDetail NavsplitviewTargetKind = 3
)

// QueryKey returns the key of the window values, which selects the view of the column for the id, like
// navsplitview.TargetKind.QueryKey of nago.
//
//gostrict:host
func (k NavsplitviewTargetKind) QueryKey(id string) string {
	return navsplitview.TargetKind(k).QueryKey(id)
}

// TNavsplitviewListItem is an item of a list, which selects a view of a split view, see NavsplitviewListItem.
type TNavsplitviewListItem struct{ ref js.Ref }

func (TNavsplitviewListItem) isView() {}

// NavsplitviewListItem returns an item, which selects the target view of the column kind, when the user clicks it,
// like navsplitview.ListItem of nago.
//
//gostrict:host
func NavsplitviewListItem(kind NavsplitviewTargetKind, target NavsplitviewViewID, content View) TNavsplitviewListItem {
	return TNavsplitviewListItem{ref: wrap(navsplitview.ListItem(navsplitview.TargetKind(kind), navsplitview.ViewID(target), viewOf(content)))}
}

// Frame sets the frame of the item.
//
//gostrict:host
func (c TNavsplitviewListItem) Frame(frame Frame) TNavsplitviewListItem {
	return TNavsplitviewListItem{ref: wrap(get[navsplitview.TListItem](c.ref).Frame(frame.nago()))}
}

// Prefix sets the prefix of the id, which the window values use for the columns.
//
//gostrict:host
func (c TNavsplitviewListItem) Prefix(id string) TNavsplitviewListItem {
	return TNavsplitviewListItem{ref: wrap(get[navsplitview.TListItem](c.ref).Prefix(id))}
}

// DeleteTarget deletes the view of the column kind, when the user selects the item.
//
//gostrict:host
func (c TNavsplitviewListItem) DeleteTarget(kind NavsplitviewTargetKind) TNavsplitviewListItem {
	return TNavsplitviewListItem{ref: wrap(get[navsplitview.TListItem](c.ref).DeleteTarget(navsplitview.TargetKind(kind)))}
}

// TNavsplitviewTwoColumn is a layout of two columns, which show the content and the detail, see
// NavsplitviewTwoColumn.
type TNavsplitviewTwoColumn struct{ ref js.Ref }

func (TNavsplitviewTwoColumn) isView() {}

// NavsplitviewTwoColumn returns a layout of two columns, whose views nav creates for each view id, like
// navsplitview.TwoColumn of nago. The content column is centered and the detail column starts at the top.
//
//gostrict:host
func NavsplitviewTwoColumn(nav func(id NavsplitviewViewID) View) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(navsplitview.TwoColumn(c6Factory(nav)))}
}

// ID sets the id, which the window values use for the columns.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) ID(id string) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).ID(id))}
}

// Frame sets the frame of the layout.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) Frame(frame Frame) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).Frame(frame.nago()))}
}

// Default sets the views, which are shown, if the window values do not select any.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) Default(content, detail NavsplitviewViewID) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).Default(navsplitview.ViewID(content), navsplitview.ViewID(detail)))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) FullWidth() TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).FullWidth())}
}

// WidthContent sets the width of the content column.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) WidthContent(width Length) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).WidthContent(nui.Length(width)))}
}

// WidthDetail sets the width of the detail column.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) WidthDetail(width Length) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).WidthDetail(nui.Length(width)))}
}

// BackgroundColorContent sets the background color of the content column.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) BackgroundColorContent(bg Color) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).BackgroundColorContent(nui.Color(bg)))}
}

// BackgroundColorDetail sets the background color of the detail column.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) BackgroundColorDetail(bg Color) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).BackgroundColorDetail(nui.Color(bg)))}
}

// AlignmentContent sets the alignment of the content column.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) AlignmentContent(alignment Alignment) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).AlignmentContent(nui.Alignment(alignment)))}
}

// AlignmentDetail sets the alignment of the detail column.
//
//gostrict:host
func (c TNavsplitviewTwoColumn) AlignmentDetail(alignment Alignment) TNavsplitviewTwoColumn {
	return TNavsplitviewTwoColumn{ref: wrap(get[navsplitview.TTwoColumn](c.ref).AlignmentDetail(nui.Alignment(alignment)))}
}

// TNavsplitviewThreeColumn is a layout of three columns, which show the sidebar, the content and the detail, see
// NavsplitviewThreeColumn.
type TNavsplitviewThreeColumn struct{ ref js.Ref }

func (TNavsplitviewThreeColumn) isView() {}

// NavsplitviewThreeColumn returns a layout of three columns, whose views nav creates for each view id, like
// navsplitview.ThreeColumn of nago. The order from left to right is sidebar, content and detail.
//
//gostrict:host
func NavsplitviewThreeColumn(nav func(id NavsplitviewViewID) View) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(navsplitview.ThreeColumn(c6Factory(nav)))}
}

// ID sets the id, which the window values use for the columns.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) ID(id string) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).ID(id))}
}

// Frame sets the frame of the layout.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) Frame(frame Frame) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).Frame(frame.nago()))}
}

// Default sets the views, which are shown, if the window values do not select any.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) Default(sidebar, content, detail NavsplitviewViewID) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).Default(navsplitview.ViewID(sidebar), navsplitview.ViewID(content), navsplitview.ViewID(detail)))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) FullWidth() TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).FullWidth())}
}

// WidthContent sets the width of the content column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) WidthContent(width Length) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).WidthContent(nui.Length(width)))}
}

// WidthSidebar sets the width of the sidebar column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) WidthSidebar(width Length) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).WidthSidebar(nui.Length(width)))}
}

// WidthDetail sets the width of the detail column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) WidthDetail(width Length) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).WidthDetail(nui.Length(width)))}
}

// BackgroundColorContent sets the background color of the content column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) BackgroundColorContent(bg Color) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).BackgroundColorContent(nui.Color(bg)))}
}

// BackgroundColorDetail sets the background color of the detail column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) BackgroundColorDetail(bg Color) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).BackgroundColorDetail(nui.Color(bg)))}
}

// BackgroundColorSidebar sets the background color of the sidebar column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) BackgroundColorSidebar(bg Color) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).BackgroundColorSidebar(nui.Color(bg)))}
}

// AlignmentContent sets the alignment of the content column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) AlignmentContent(alignment Alignment) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).AlignmentContent(nui.Alignment(alignment)))}
}

// AlignmentDetail sets the alignment of the detail column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) AlignmentDetail(alignment Alignment) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).AlignmentDetail(nui.Alignment(alignment)))}
}

// AlignmentSidebar sets the alignment of the sidebar column.
//
//gostrict:host
func (c TNavsplitviewThreeColumn) AlignmentSidebar(alignment Alignment) TNavsplitviewThreeColumn {
	return TNavsplitviewThreeColumn{ref: wrap(get[navsplitview.TThreeColumn](c.ref).AlignmentSidebar(nui.Alignment(alignment)))}
}

// NavsplitviewNavigateContent shows the view in the content column of the split view with the id, like
// navsplitview.NavigateContent of nago.
//
//gostrict:host
func NavsplitviewNavigateContent(wnd Window, id string, view NavsplitviewViewID) {
	navsplitview.NavigateContent(wnd.nago(), id, navsplitview.ViewID(view))
}

// NavsplitviewNavigateDetail shows the view in the detail column of the split view with the id, like
// navsplitview.NavigateDetail of nago.
//
//gostrict:host
func NavsplitviewNavigateDetail(wnd Window, id string, view NavsplitviewViewID) {
	navsplitview.NavigateDetail(wnd.nago(), id, navsplitview.ViewID(view))
}

// NavsplitviewNavigateSidebar shows the view in the sidebar column of the split view with the id, like
// navsplitview.NavigateSidebar of nago.
//
//gostrict:host
func NavsplitviewNavigateSidebar(wnd Window, id string, view NavsplitviewViewID) {
	navsplitview.NavigateSidebar(wnd.nago(), id, navsplitview.ViewID(view))
}
