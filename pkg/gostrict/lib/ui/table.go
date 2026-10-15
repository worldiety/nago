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

// TTableColumn is a column of a table, see TableColumn. It defines the header of the column and the defaults of
// its cells.
type TTableColumn struct{ ref js.Ref }

// TableColumn returns a column with the header content. Its cells are aligned at the start by default.
//
//gostrict:host
func TableColumn(content View) TTableColumn {
	return TTableColumn{ref: wrap(nui.TableColumn(viewOf(content)))}
}

// Content sets the header content.
//
//gostrict:host
func (c TTableColumn) Content(content View) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).Content(viewOf(content)))}
}

// Action calls action, when the user clicks a cell of the column.
//
//gostrict:host
func (c TTableColumn) Action(action func()) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).Action(action))}
}

// Key identifies the action across renders by what it acts on, e.g. ("sort", column).
//
//gostrict:host
func (c TTableColumn) Key(name, id string) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).Key(name, id))}
}

// HoveredBackgroundColor sets the background color of the cells, while they are hovered.
//
//gostrict:host
func (c TTableColumn) HoveredBackgroundColor(backgroundColor Color) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).HoveredBackgroundColor(nui.Color(backgroundColor)))}
}

// Width sets the width of the column.
//
//gostrict:host
func (c TTableColumn) Width(width Length) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).Width(nui.Length(width)))}
}

// Alignment aligns the content of the cells.
//
//gostrict:host
func (c TTableColumn) Alignment(alignment Alignment) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).Alignment(nui.Alignment(alignment)))}
}

// BackgroundColor sets the background color of the cells.
//
//gostrict:host
func (c TTableColumn) BackgroundColor(backgroundColor Color) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).BackgroundColor(nui.Color(backgroundColor)))}
}

// Padding sets the padding of the cells.
//
//gostrict:host
func (c TTableColumn) Padding(padding Padding) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).Padding(padding.nago()))}
}

// Border sets the border of the cells.
//
//gostrict:host
func (c TTableColumn) Border(border Border) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).Border(border.nago()))}
}

// Span sets how many columns the header spans.
//
//gostrict:host
func (c TTableColumn) Span(span int32) TTableColumn {
	return TTableColumn{ref: wrap(get[nui.TTableColumn](c.ref).Span(int(span)))}
}

// TTableCell is a cell of a row, see TableCell.
type TTableCell struct{ ref js.Ref }

// TableCell returns a cell with the content.
//
//gostrict:host
func TableCell(content View) TTableCell {
	return TTableCell{ref: wrap(nui.TableCell(viewOf(content)))}
}

// Content sets the content.
//
//gostrict:host
func (c TTableCell) Content(content View) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).Content(viewOf(content)))}
}

// ColSpan sets how many columns the cell spans.
//
//gostrict:host
func (c TTableCell) ColSpan(colSpan int32) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).ColSpan(int(colSpan)))}
}

// Action calls action, when the user clicks the cell.
//
//gostrict:host
func (c TTableCell) Action(action func()) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).Action(action))}
}

// Key identifies the action across renders by what it acts on, e.g. ("column", rowID).
//
//gostrict:host
func (c TTableCell) Key(name, id string) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).Key(name, id))}
}

// RowSpan sets how many rows the cell spans.
//
//gostrict:host
func (c TTableCell) RowSpan(rowSpan int32) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).RowSpan(int(rowSpan)))}
}

// Alignment aligns the content of the cell.
//
//gostrict:host
func (c TTableCell) Alignment(alignment Alignment) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).Alignment(nui.Alignment(alignment)))}
}

// BackgroundColor sets the background color of the cell.
//
//gostrict:host
func (c TTableCell) BackgroundColor(backgroundColor Color) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).BackgroundColor(nui.Color(backgroundColor)))}
}

// HoveredBackgroundColor sets the background color of the cell, while it is hovered.
//
//gostrict:host
func (c TTableCell) HoveredBackgroundColor(backgroundColor Color) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).HoveredBackgroundColor(nui.Color(backgroundColor)))}
}

// Padding sets the padding of the cell.
//
//gostrict:host
func (c TTableCell) Padding(padding Padding) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).Padding(padding.nago()))}
}

// Border sets the border of the cell.
//
//gostrict:host
func (c TTableCell) Border(border Border) TTableCell {
	return TTableCell{ref: wrap(get[nui.TTableCell](c.ref).Border(border.nago()))}
}

// TTableRow is a row of a table, see TableRow.
type TTableRow struct{ ref js.Ref }

// TableRow returns a row with the cells.
//
//gostrict:host
func TableRow(cells ...TTableCell) TTableRow {
	return TTableRow{ref: wrap(nui.TableRow(b4CellsOf(cells)...))}
}

// Append appends the cells.
//
//gostrict:host
func (r TTableRow) Append(cells ...TTableCell) TTableRow {
	return TTableRow{ref: wrap(get[nui.TTableRow](r.ref).Append(b4CellsOf(cells)...))}
}

// Action calls action, when the user clicks the row.
//
//gostrict:host
func (r TTableRow) Action(action func()) TTableRow {
	return TTableRow{ref: wrap(get[nui.TTableRow](r.ref).Action(action))}
}

// Key identifies the action across renders by what it acts on, e.g. ("row", rowID).
//
//gostrict:host
func (r TTableRow) Key(name, id string) TTableRow {
	return TTableRow{ref: wrap(get[nui.TTableRow](r.ref).Key(name, id))}
}

// Height sets the height of the row.
//
//gostrict:host
func (r TTableRow) Height(height Length) TTableRow {
	return TTableRow{ref: wrap(get[nui.TTableRow](r.ref).Height(nui.Length(height)))}
}

// BackgroundColor sets the background color of the row.
//
//gostrict:host
func (r TTableRow) BackgroundColor(backgroundColor Color) TTableRow {
	return TTableRow{ref: wrap(get[nui.TTableRow](r.ref).BackgroundColor(nui.Color(backgroundColor)))}
}

// HoveredBackgroundColor sets the background color of the row, while it is hovered.
//
//gostrict:host
func (r TTableRow) HoveredBackgroundColor(backgroundColor Color) TTableRow {
	return TTableRow{ref: wrap(get[nui.TTableRow](r.ref).HoveredBackgroundColor(nui.Color(backgroundColor)))}
}

// TTable is a table with a header, rows, borders and background, see Table.
type TTable struct{ ref js.Ref }

func (TTable) isView() {}

// Table returns a table with the columns and the default styling.
//
//gostrict:host
func Table(columns ...TTableColumn) TTable {
	return TTable{ref: wrap(nui.Table(b4ColumnsOf(columns)...))}
}

// Columns sets the columns of the header.
//
//gostrict:host
func (c TTable) Columns(columns ...TTableColumn) TTable {
	return TTable{ref: wrap(get[nui.TTable](c.ref).Columns(b4ColumnsOf(columns)...))}
}

// BackgroundColor sets the background color of the table.
//
//gostrict:host
func (c TTable) BackgroundColor(backgroundColor Color) TTable {
	return TTable{ref: wrap(get[nui.TTable](c.ref).BackgroundColor(nui.Color(backgroundColor)))}
}

// Border sets the border of the table.
//
//gostrict:host
func (c TTable) Border(border Border) TTable {
	return TTable{ref: wrap(get[nui.TTable](c.ref).Border(border.nago()))}
}

// Frame sets the size of the table.
//
//gostrict:host
func (c TTable) Frame(frame Frame) TTable {
	return TTable{ref: wrap(get[nui.TTable](c.ref).Frame(frame.nago()))}
}

// RowDividerColor sets the color of the dividers between the rows.
//
//gostrict:host
func (c TTable) RowDividerColor(color Color) TTable {
	return TTable{ref: wrap(get[nui.TTable](c.ref).RowDividerColor(nui.Color(color)))}
}

// HeaderDividerColor sets the color of the divider between the header and the body.
//
//gostrict:host
func (c TTable) HeaderDividerColor(color Color) TTable {
	return TTable{ref: wrap(get[nui.TTable](c.ref).HeaderDividerColor(nui.Color(color)))}
}

// Rows appends the rows.
//
//gostrict:host
func (c TTable) Rows(rows ...TTableRow) TTable {
	return TTable{ref: wrap(get[nui.TTable](c.ref).Rows(b4RowsOf(rows)...))}
}

// CellPadding sets the default padding of all cells. The padding of a cell overrides it.
//
//gostrict:host
func (c TTable) CellPadding(padding Padding) TTable {
	return TTable{ref: wrap(get[nui.TTable](c.ref).CellPadding(padding.nago()))}
}
