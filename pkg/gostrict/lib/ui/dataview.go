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
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

// DataViewStyle is the style of a data view, like dataview.Style of nago.
type DataViewStyle int32

const (
	// DataViewAuto chooses a table on a large window and cards or a list otherwise.
	DataViewAuto  DataViewStyle = 0
	DataViewTable DataViewStyle = 1
	DataViewCard  DataViewStyle = 2
	DataViewList  DataViewStyle = 3
)

// DataViewFormatHint tells, how a card or a list shows a field, like dataview.FormatHint of nago.
type DataViewFormatHint int32

const (
	DataViewHintInline    DataViewFormatHint = 1
	DataViewHintInvisible DataViewFormatHint = 2
)

// TDataView shows rows of the program as a table, cards or a list with search, sorting, selection and actions, like
// dataview.TDataView of nago. The program keeps the rows and identifies them by their index from 0 to rows-1, so
// that no repository is involved, see DataViewFromSlice.
type TDataView struct{ ref js.Ref }

func (TDataView) isView() {}

// TDataViewField is a column of a data view, like dataview.Field of nago, see DataViewField.
type TDataViewField struct{ ref js.Ref }

// TDataViewSelectOption is an action on the selected rows, like dataview.SelectOption of nago.
type TDataViewSelectOption struct{ ref js.Ref }

// TDataViewCreateOption is an entry of the menu, which creates a row, like dataview.CreateOption of nago.
type TDataViewCreateOption struct{ ref js.Ref }

// DataViewFromSlice returns a data view of rows, which the program keeps, e.g. in a list.List, like
// dataview.FromSlice of nago. The fields show each row by its index:
//
//	ui.DataViewFromSlice(wnd, people.Len(),
//		ui.DataViewField("Name", func(row int32) ui.View { p, _ := people.Get(row); return ui.Text(p.Name) }),
//	).Action(func(row int32) { open(row) })
//
//gostrict:host
func DataViewFromSlice(wnd Window, rows int32, fields ...TDataViewField) TDataView {
	return TDataView{ref: wrap(newDataView(wnd.nago(), rows, fields))}
}

// DataViewField returns a field, which shows the view of each row.
//
//gostrict:host
func DataViewField(name string, view func(row int32) View) TDataViewField {
	return TDataViewField{ref: wrap(dataViewField{name: name, view: view})}
}

// ID sets the identifier of the field, e.g. for CardTitle or ColumnWidth.
//
//gostrict:host
func (f TDataViewField) ID(id string) TDataViewField {
	return f.with(func(d *dataViewField) { d.id = id })
}

// Comparator lets the user sort the rows by the field: cmp returns a negative number, if row a comes first, 0, if
// both are equal, and a positive number otherwise.
//
//gostrict:host
func (f TDataViewField) Comparator(cmp func(a, b int32) int32) TDataViewField {
	return f.with(func(d *dataViewField) { d.cmp = cmp })
}

// MinSize shows the field only on windows of the size class or larger, like dataview.MinSizeMedium of nago.
//
//gostrict:host
func (f TDataViewField) MinSize(class WindowSizeClass) TDataViewField {
	return f.with(func(d *dataViewField) { d.minSize, d.hasMinSize = class, true })
}

// DataViewSelectOption returns an action on the selected rows, which the user finds in the bar of the selection.
//
//gostrict:host
func DataViewSelectOption(icon SVG, name string, action func(rows list.List[int32]) error) TDataViewSelectOption {
	return TDataViewSelectOption{ref: wrap(dataViewSelect{icon: icon, name: name, action: action})}
}

// DataViewSelectOptionDelete returns the action, which deletes the selected rows after a standard confirmation, like
// dataview.NewSelectOptionDelete of nago.
//
//gostrict:host
func DataViewSelectOptionDelete(wnd Window, action func(rows list.List[int32]) error) TDataViewSelectOption {
	return TDataViewSelectOption{ref: wrap(dataViewSelect{delete: true, wnd: wnd.nago(), action: action})}
}

// Visible shows the action only, if visible returns true for the selected rows.
//
//gostrict:host
func (o TDataViewSelectOption) Visible(visible func(rows list.List[int32]) bool) TDataViewSelectOption {
	return o.with(func(d *dataViewSelect) { d.visible = visible })
}

// ConfirmDialog asks the user with the title and the message, before the action runs.
//
//gostrict:host
func (o TDataViewSelectOption) ConfirmDialog(title, message string) TDataViewSelectOption {
	return o.with(func(d *dataViewSelect) { d.confirmTitle, d.confirmMessage = title, message })
}

// DataViewCreateOption returns an entry of the menu, which creates a row.
//
//gostrict:host
func DataViewCreateOption(icon SVG, name string, action func() error) TDataViewCreateOption {
	return TDataViewCreateOption{ref: wrap(dataViewCreate{icon: icon, name: name, action: action})}
}

// Action calls fn with the row, which the user clicks.
//
//gostrict:host
func (c TDataView) Action(fn func(row int32)) TDataView {
	return c.with(func(d *dataView) { d.action = fn })
}

// NextActionIndicator shows a chevron at each row, which hints at its action.
//
//gostrict:host
func (c TDataView) NextActionIndicator(b bool) TDataView {
	return c.with(func(d *dataView) { d.nextIndicator = b })
}

// CreateAction shows a button, which calls fn to create a row.
//
//gostrict:host
func (c TDataView) CreateAction(fn func()) TDataView {
	return c.with(func(d *dataView) { d.create = fn })
}

// CreateOptions shows a menu of the options, which create a row.
//
//gostrict:host
func (c TDataView) CreateOptions(options ...TDataViewCreateOption) TDataView {
	return c.with(func(d *dataView) { d.createOptions = options })
}

// CreateActionView shows the view instead of the button, which creates a row.
//
//gostrict:host
func (c TDataView) CreateActionView(view View) TDataView {
	return c.with(func(d *dataView) { d.createView = viewOf(view) })
}

// Selection shows or hides the checkboxes, which select rows.
//
//gostrict:host
func (c TDataView) Selection(show bool) TDataView {
	return c.with(func(d *dataView) { d.hideSelection = !show })
}

// SelectOptions sets the actions on the selected rows.
//
//gostrict:host
func (c TDataView) SelectOptions(options ...TDataViewSelectOption) TDataView {
	return c.with(func(d *dataView) { d.selectOptions = options })
}

// SelectionChanged calls fn with the selected rows, whenever the selection changes.
//
//gostrict:host
func (c TDataView) SelectionChanged(fn func(rows list.List[int32])) TDataView {
	return c.with(func(d *dataView) { d.selectionChanged = fn })
}

// Style sets the style.
//
//gostrict:host
func (c TDataView) Style(style DataViewStyle) TDataView {
	return c.with(func(d *dataView) { d.style = style })
}

// Search shows or hides the field, which searches the rows.
//
//gostrict:host
func (c TDataView) Search(visible bool) TDataView {
	return c.with(func(d *dataView) { d.search = visible })
}

// ActionBarContent shows the view in the bar above the rows.
//
//gostrict:host
func (c TDataView) ActionBarContent(content View) TDataView {
	return c.with(func(d *dataView) { d.actionBar = viewOf(content) })
}

// CardTitle shows the field with the id as the title of a card, see TDataViewField.ID.
//
//gostrict:host
func (c TDataView) CardTitle(field string) TDataView {
	return c.with(func(d *dataView) { d.cardTitle = field })
}

// CardHint sets, how a card shows the field with the id.
//
//gostrict:host
func (c TDataView) CardHint(field string, hint DataViewFormatHint) TDataView {
	return c.with(func(d *dataView) { d.cardHints = withHint(d.cardHints, field, hint) })
}

// ListTitle and ListDescription show the fields with the ids as the title and the description of an entry of the
// list.
//
//gostrict:host
func (c TDataView) ListTitle(field string) TDataView {
	return c.with(func(d *dataView) { d.listTitle = field })
}

// ListDescription shows the field with the id as the description of an entry of the list.
//
//gostrict:host
func (c TDataView) ListDescription(field string) TDataView {
	return c.with(func(d *dataView) { d.listDescription = field })
}

// ListHint sets, how an entry of the list shows the field with the id.
//
//gostrict:host
func (c TDataView) ListHint(field string, hint DataViewFormatHint) TDataView {
	return c.with(func(d *dataView) { d.listHints = withHint(d.listHints, field, hint) })
}

// ListHighlight highlights the row in the list.
//
//gostrict:host
func (c TDataView) ListHighlight(row int32) TDataView {
	return c.with(func(d *dataView) { d.highlight = append(d.highlight[:len(d.highlight):len(d.highlight)], row) })
}

// ColumnWidth sets the width of the column of the field with the id in a table.
//
//gostrict:host
func (c TDataView) ColumnWidth(field string, width Length) TDataView {
	return c.with(func(d *dataView) { d.widths = withWidth(d.widths, field, width) })
}
