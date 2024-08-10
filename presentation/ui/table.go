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

func (c TTableColumn) HoveredBackgroundColor(backgroundColor Color) TTableColumn {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

func (c TTableColumn) Width(width Length) TTableColumn {
	c.width = width.ora()
	return c
}

func (c TTableColumn) Alignment(alignment Alignment) TTableColumn {
	c.alignment = alignment.ora()
	return c
}

func (c TTableColumn) BackgroundColor(backgroundColor Color) TTableColumn {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TTableColumn) Padding(padding Padding) TTableColumn {
	c.padding = padding.ora()
	return c
}

func (c TTableColumn) Border(border Border) TTableColumn {
	c.border = border.ora()
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

func (c TTableCell) Alignment(alignment Alignment) TTableCell {
	c.alignment = alignment.ora()
	return c
}

func (c TTableCell) BackgroundColor(backgroundColor Color) TTableCell {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TTableCell) HoveredBackgroundColor(backgroundColor Color) TTableCell {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

func (c TTableCell) Padding(padding Padding) TTableCell {
	c.padding = padding.ora()
	return c
}

func (c TTableCell) Border(border Border) TTableCell {
	c.border = border.ora()
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

func (r TTableRow) Height(height Length) TTableRow {
	r.height = height.ora()
	return r
}

func (r TTableRow) BackgroundColor(backgroundColor Color) TTableRow {
	r.backgroundColor = backgroundColor.ora()
	return r
}

func (c TTableRow) HoveredBackgroundColor(backgroundColor Color) TTableRow {
	c.hoveredBackgroundColor = backgroundColor.ora()
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
	return TTable{
		columns:             columns,
		backgroundColor:     M2.ora(),
		defaultCellPaddings: Padding{}.Horizontal(L24).Vertical(L16).ora(),
		rowDividerColor:     M5.ora(),
		border:              Border{}.Radius(L20).ora(),
	}
}

func (c TTable) BackgroundColor(backgroundColor Color) TTable {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TTable) Border(border Border) TTable {
	c.border = border.ora()
	return c
}

func (c TTable) Frame(frame Frame) TTable {
	c.frame = frame.ora()
	return c
}

func (c TTable) RowDividerColor(color Color) TTable {
	c.rowDividerColor = color.ora()
	return c
}

func (c TTable) Rows(rows ...TTableRow) TTable {
	c.rows = rows
	return c
}

// CellPadding sets the default cell padding for all cells at once.
// Individual cell paddings take precedence.
func (c TTable) CellPadding(padding Padding) TTable {
	c.defaultCellPaddings = padding.ora()
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
