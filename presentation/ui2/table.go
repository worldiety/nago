package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TTableColumn struct {
	content                core.View
	colSpan                int
	width                  ora.Length
	alignment              ora.Alignment
	backgroundColor        ora.Color
	hoveredBackgroundColor ora.Color
	padding                ora.Padding
	border                 ora.Border
	action                 func()
}

func TableColumn(content core.View) TTableColumn {
	return TTableColumn{content: content}
}

// Action refers only to the cell, not to the entire column.
func (c TTableColumn) Action(action func()) TTableColumn {
	c.action = action
	return c
}

func (c TTableColumn) HoveredBackgroundColor(backgroundColor ora.Color) TTableColumn {
	c.hoveredBackgroundColor = backgroundColor
	return c
}

func (c TTableColumn) Width(width ora.Length) TTableColumn {
	c.width = width
	return c
}

func (c TTableColumn) Alignment(alignment ora.Alignment) TTableColumn {
	c.alignment = alignment
	return c
}

func (c TTableColumn) BackgroundColor(backgroundColor ora.Color) TTableColumn {
	c.backgroundColor = backgroundColor
	return c
}

func (c TTableColumn) Padding(padding ora.Padding) TTableColumn {
	c.padding = padding
	return c
}

func (c TTableColumn) Border(border ora.Border) TTableColumn {
	c.border = border
	return c
}

func (c TTableColumn) Span(span int) TTableColumn {
	c.colSpan = span
	return c
}

//

type TTableCell struct {
	content                core.View
	colSpan                int
	rowSpan                int
	alignment              ora.Alignment
	backgroundColor        ora.Color
	hoveredBackgroundColor ora.Color
	padding                ora.Padding
	border                 ora.Border
	action                 func()
}

func TableCell(content core.View) TTableCell {
	return TTableCell{content: content}
}

func (c TTableCell) ColSpan(colSpan int) TTableCell {
	c.colSpan = colSpan
	return c
}

func (c TTableCell) Action(action func()) TTableCell {
	c.action = action
	return c
}

func (c TTableCell) RowSpan(rowSpan int) TTableCell {
	c.rowSpan = rowSpan
	return c
}

func (c TTableCell) Alignment(alignment ora.Alignment) TTableCell {
	c.alignment = alignment
	return c
}

func (c TTableCell) BackgroundColor(backgroundColor ora.Color) TTableCell {
	c.backgroundColor = backgroundColor
	return c
}

func (c TTableCell) HoveredBackgroundColor(backgroundColor ora.Color) TTableCell {
	c.hoveredBackgroundColor = backgroundColor
	return c
}

func (c TTableCell) Padding(padding ora.Padding) TTableCell {
	c.padding = padding
	return c
}

func (c TTableCell) Border(border ora.Border) TTableCell {
	c.border = border
	return c
}

//

type TTableRow struct {
	cells                  []TTableCell
	height                 ora.Length
	backgroundColor        ora.Color
	hoveredBackgroundColor ora.Color
	action                 func()
}

func TableRow(cells ...TTableCell) TTableRow {
	return TTableRow{cells: cells}
}

func (r TTableRow) Action(action func()) TTableRow {
	r.action = action
	return r
}

func (r TTableRow) Height(height ora.Length) TTableRow {
	r.height = height
	return r
}

func (r TTableRow) BackgroundColor(backgroundColor ora.Color) TTableRow {
	r.backgroundColor = backgroundColor
	return r
}

func (c TTableRow) HoveredBackgroundColor(backgroundColor ora.Color) TTableRow {
	c.hoveredBackgroundColor = backgroundColor
	return c
}

type TTable struct {
	columns             []TTableColumn
	rows                []TTableRow
	frame               ora.Frame
	border              ora.Border
	backgroundColor     ora.Color
	defaultCellPaddings ora.Padding
	rowDividerColor     ora.Color
}

func Table(columns ...TTableColumn) TTable {
	return TTable{columns: columns}
}

func (c TTable) BackgroundColor(backgroundColor ora.Color) TTable {
	c.backgroundColor = backgroundColor
	return c
}

func (c TTable) Border(border ora.Border) TTable {
	c.border = border
	return c
}

func (c TTable) Frame(frame ora.Frame) TTable {
	c.frame = frame
	return c
}

func (c TTable) RowDividerColor(color ora.Color) TTable {
	c.rowDividerColor = color
	return c
}

func (c TTable) Rows(rows ...TTableRow) TTable {
	c.rows = rows
	return c
}

// CellPadding sets the default cell padding for all cells at once.
// Individual cell paddings take precedence.
func (c TTable) CellPadding(padding ora.Padding) TTable {
	c.defaultCellPaddings = padding
	return c
}

func (c TTable) Render(ctx core.RenderContext) ora.Component {
	var header ora.TableHeader
	for _, column := range c.columns {
		header.Columns = append(header.Columns, ora.TableColumn{
			Content:                    render(ctx, column.content),
			ColSpan:                    column.colSpan,
			Width:                      column.width,
			Alignment:                  column.alignment,
			CellBackgroundColor:        column.backgroundColor,
			CellAction:                 ctx.MountCallback(column.action),
			CellPadding:                column.padding,
			CellBorder:                 column.border,
			CellHoveredBackgroundColor: column.hoveredBackgroundColor,
		})
	}

	rows := make([]ora.TableRow, 0, len(c.rows))
	for _, row := range c.rows {
		cells := make([]ora.TableCell, 0, len(row.cells))
		for _, cell := range row.cells {
			cells = append(cells, ora.TableCell{
				Content:                render(ctx, cell.content),
				RowSpan:                cell.rowSpan,
				ColSpan:                cell.colSpan,
				Alignment:              cell.alignment,
				BackgroundColor:        cell.backgroundColor,
				Border:                 cell.border,
				Action:                 ctx.MountCallback(cell.action),
				HoveredBackgroundColor: cell.hoveredBackgroundColor,
			})
		}

		rows = append(rows, ora.TableRow{
			Cells:                  cells,
			Height:                 row.height,
			BackgroundColor:        row.backgroundColor,
			HoveredBackgroundColor: row.hoveredBackgroundColor,
			Action:                 ctx.MountCallback(row.action),
		})
	}

	return ora.Table{
		Type:               ora.TableT,
		Header:             header,
		Rows:               rows,
		Frame:              c.frame,
		Border:             c.border,
		BackgroundColor:    c.backgroundColor,
		DefaultCellPadding: c.defaultCellPaddings,
		RowDividerColor:    c.rowDividerColor,
	}
}
