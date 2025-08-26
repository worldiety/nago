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

// TTableColumn is a layout component (Table Column).
// It defines the configuration for a table column header and its cell defaults.
// Columns can define width, alignment, background color, padding, borders,
// and cell-specific actions (e.g., sorting).
type TTableColumn struct {
	content                core.View       // header content
	colSpan                int             // number of columns to span
	width                  proto.Length    // column width
	alignment              proto.Alignment // content alignment
	backgroundColor        proto.Color     // background color for cells
	hoveredBackgroundColor proto.Color     // background color on hover
	padding                proto.Padding   // padding inside the column cell
	border                 proto.Border    // border around the cell
	action                 func()          // optional cell-specific action
}

// TableColumn creates a new table column with the given header content.
func TableColumn(content core.View) TTableColumn {
	return TTableColumn{
		alignment: proto.Leading, // a leading start is more common in standard tables
		content:   content,
	}
}

// Action sets an optional click/tap action for the column's cells.
func (c TTableColumn) Action(action func()) TTableColumn {
	c.action = action
	return c
}

// HoveredBackgroundColor sets the background color when the column cell is hovered.
func (c TTableColumn) HoveredBackgroundColor(backgroundColor Color) TTableColumn {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

// Width sets the column width.
func (c TTableColumn) Width(width Length) TTableColumn {
	c.width = width.ora()
	return c
}

// Alignment sets the content alignment within the column cell.
func (c TTableColumn) Alignment(alignment Alignment) TTableColumn {
	c.alignment = alignment.ora()
	return c
}

// BackgroundColor sets the background color for the column cell.
func (c TTableColumn) BackgroundColor(backgroundColor Color) TTableColumn {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// Padding sets the padding for the column cell.
func (c TTableColumn) Padding(padding Padding) TTableColumn {
	c.padding = padding.ora()
	return c
}

// Border sets the border for the column cell.
func (c TTableColumn) Border(border Border) TTableColumn {
	c.border = border.ora()
	return c
}

// Span sets how many columns this header should span.
func (c TTableColumn) Span(span int) TTableColumn {
	c.colSpan = span
	return c
}

//

// TTableCell is a layout component (Table Cell).
// Represents an individual cell inside a table row with optional spanning,
// alignment, background, padding, border, and actions.
type TTableCell struct {
	content                core.View
	colSpan                int
	rowSpan                int
	alignment              proto.Alignment
	backgroundColor        proto.Color
	hoveredBackgroundColor proto.Color
	padding                proto.Padding
	border                 proto.Border
	action                 func()
}

// TableCell creates a new table cell with the given content.
func TableCell(content core.View) TTableCell {
	return TTableCell{content: content}
}

// ColSpan sets how many columns this cell spans.
func (c TTableCell) ColSpan(colSpan int) TTableCell {
	c.colSpan = colSpan
	return c
}

// Action sets an optional click/tap action for the cell.
func (c TTableCell) Action(action func()) TTableCell {
	c.action = action
	return c
}

// RowSpan sets how many rows this cell spans.
func (c TTableCell) RowSpan(rowSpan int) TTableCell {
	c.rowSpan = rowSpan
	return c
}

// Alignment sets the alignment for the cell content.
func (c TTableCell) Alignment(alignment Alignment) TTableCell {
	c.alignment = alignment.ora()
	return c
}

// BackgroundColor sets the background color of the cell.
func (c TTableCell) BackgroundColor(backgroundColor Color) TTableCell {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// HoveredBackgroundColor sets the background color when the cell is hovered.
func (c TTableCell) HoveredBackgroundColor(backgroundColor Color) TTableCell {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

// Padding sets the padding of the cell.
func (c TTableCell) Padding(padding Padding) TTableCell {
	c.padding = padding.ora()
	return c
}

// Border sets the border of the cell.
func (c TTableCell) Border(border Border) TTableCell {
	c.border = border.ora()
	return c
}

//

// TTableRow is a layout component (Table Row).
// It groups a collection of cells and defines row-level styling and actions.
type TTableRow struct {
	cells                  []TTableCell
	height                 proto.Length
	backgroundColor        proto.Color
	hoveredBackgroundColor proto.Color
	action                 func()
}

// TableRow creates a new table row with the given cells.
func TableRow(cells ...TTableCell) TTableRow {
	return TTableRow{cells: cells}
}

// Action sets a click/tap action for the entire row.
func (r TTableRow) Action(action func()) TTableRow {
	r.action = action
	return r
}

// Height sets the row height.
func (r TTableRow) Height(height Length) TTableRow {
	r.height = height.ora()
	return r
}

// BackgroundColor sets the background color of the row.
func (r TTableRow) BackgroundColor(backgroundColor Color) TTableRow {
	r.backgroundColor = backgroundColor.ora()
	return r
}

// HoveredBackgroundColor sets the background color when the row is hovered.
func (c TTableRow) HoveredBackgroundColor(backgroundColor Color) TTableRow {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

// TTable is a composite component (Table).
// It represents a full table structure with header, rows, borders,
// background styling, and default cell paddings.
type TTable struct {
	columns             []TTableColumn
	rows                []TTableRow
	frame               proto.Frame
	border              proto.Border
	backgroundColor     proto.Color
	defaultCellPaddings proto.Padding
	rowDividerColor     proto.Color
	headerDividerColor  proto.Color
}

// Table creates a new table with the specified columns and default styling.
func Table(columns ...TTableColumn) TTable {
	return TTable{
		columns:             columns,
		backgroundColor:     M2.ora(),
		defaultCellPaddings: Padding{}.Horizontal(L24).Vertical(L16).ora(),
		rowDividerColor:     M5.ora(),
		border:              Border{}.Radius(L20).ora(),
	}
}

// BackgroundColor sets the background color of the table.
func (c TTable) BackgroundColor(backgroundColor Color) TTable {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// Border sets the border of the table.
func (c TTable) Border(border Border) TTable {
	c.border = border.ora()
	return c
}

// Frame sets the frame of the table.
func (c TTable) Frame(frame Frame) TTable {
	c.frame = frame.ora()
	return c
}

// RowDividerColor sets the divider color between rows.
func (c TTable) RowDividerColor(color Color) TTable {
	c.rowDividerColor = color.ora()
	return c
}

// HeaderDividerColor sets the divider color between header and body.
func (c TTable) HeaderDividerColor(color Color) TTable {
	c.headerDividerColor = color.ora()
	return c
}

// Rows appends one or more rows to the table.
func (c TTable) Rows(rows ...TTableRow) TTable {
	c.rows = append(c.rows, rows...)
	return c
}

// CellPadding sets the default cell padding for all cells.
// Specific cell padding overrides this setting.
func (c TTable) CellPadding(padding Padding) TTable {
	c.defaultCellPaddings = padding.ora()
	return c
}

// Render builds the protocol representation of the table,
// including headers, rows, styling, paddings, and dividers.
func (c TTable) Render(ctx core.RenderContext) core.RenderNode {
	var header proto.TableHeader
	for _, column := range c.columns {
		header.Columns = append(header.Columns, proto.TableColumn{
			Content:                    render(ctx, column.content),
			ColSpan:                    proto.Uint(column.colSpan),
			Width:                      column.width,
			Alignment:                  column.alignment,
			CellBackgroundColor:        column.backgroundColor,
			CellAction:                 ctx.MountCallback(column.action),
			CellPadding:                column.padding,
			CellBorder:                 column.border,
			CellHoveredBackgroundColor: column.hoveredBackgroundColor,
		})
	}

	rows := make([]proto.TableRow, 0, len(c.rows))
	for _, row := range c.rows {
		cells := make([]proto.TableCell, 0, len(row.cells))
		for _, cell := range row.cells {
			cells = append(cells, proto.TableCell{
				Content:                render(ctx, cell.content),
				RowSpan:                proto.Uint(cell.rowSpan),
				ColSpan:                proto.Uint(cell.colSpan),
				Alignment:              cell.alignment,
				BackgroundColor:        cell.backgroundColor,
				Border:                 cell.border,
				Action:                 ctx.MountCallback(cell.action),
				HoveredBackgroundColor: cell.hoveredBackgroundColor,
			})
		}

		rows = append(rows, proto.TableRow{
			Cells:                  cells,
			Height:                 row.height,
			BackgroundColor:        row.backgroundColor,
			HoveredBackgroundColor: row.hoveredBackgroundColor,
			Action:                 ctx.MountCallback(row.action),
		})
	}

	return &proto.Table{
		Header:             header,
		Rows:               rows,
		Frame:              c.frame,
		Border:             c.border,
		BackgroundColor:    c.backgroundColor,
		DefaultCellPadding: c.defaultCellPaddings,
		RowDividerColor:    c.rowDividerColor,
		HeaderDividerColor: c.headerDividerColor,
	}
}
