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

// TGridCell is a cell of a grid, see GridCell. It places its body by its rows and columns.
type TGridCell struct{ ref js.Ref }

// GridCell returns a cell with the body. Rows and columns start at 1. By default, the body stretches to fill the
// cell.
//
//gostrict:host
func GridCell(body View) TGridCell {
	return TGridCell{ref: wrap(nui.GridCell(viewOf(body)))}
}

// Padding sets the inner spacing around the content of the cell.
//
//gostrict:host
func (c TGridCell) Padding(padding Padding) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).Padding(padding.nago()))}
}

// BackgroundColor sets the background color of the cell.
//
//gostrict:host
func (c TGridCell) BackgroundColor(backgroundColor Color) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).BackgroundColor(nui.Color(backgroundColor)))}
}

// Alignment aligns the content within the cell.
//
//gostrict:host
func (c TGridCell) Alignment(alignment Alignment) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).Alignment(nui.Alignment(alignment)))}
}

// ColStart sets the starting column, which must start at 1.
//
//gostrict:host
func (c TGridCell) ColStart(colStart int32) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).ColStart(int(colStart)))}
}

// ColEnd sets the ending column, which must be at least one more than the starting column.
//
//gostrict:host
func (c TGridCell) ColEnd(colEnd int32) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).ColEnd(int(colEnd)))}
}

// RowStart sets the starting row, which must start at 1.
//
//gostrict:host
func (c TGridCell) RowStart(rowStart int32) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).RowStart(int(rowStart)))}
}

// RowEnd sets the ending row, which must be at least one more than the starting row.
//
//gostrict:host
func (c TGridCell) RowEnd(rowEnd int32) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).RowEnd(int(rowEnd)))}
}

// ColSpan sets how many columns the cell spans.
//
//gostrict:host
func (c TGridCell) ColSpan(colSpan int32) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).ColSpan(int(colSpan)))}
}

// RowSpan sets how many rows the cell spans.
//
//gostrict:host
func (c TGridCell) RowSpan(rowSpan int32) TGridCell {
	return TGridCell{ref: wrap(get[nui.TGridCell](c.ref).RowSpan(int(rowSpan)))}
}

// TGrid is a layout, which arranges its cells in rows and columns, see Grid.
type TGrid struct{ ref js.Ref }

func (TGrid) isView() {}

// Grid returns a grid with the cells. For simple cases, only the rows or the columns need to be set.
//
//gostrict:host
func Grid(cells ...TGridCell) TGrid {
	return TGrid{ref: wrap(nui.Grid(b4GridCellsOf(cells)...))}
}

// Append appends the cells.
//
//gostrict:host
func (c TGrid) Append(cells ...TGridCell) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Append(b4GridCellsOf(cells)...))}
}

// Rows sets the amount of rows explicitly.
//
//gostrict:host
func (c TGrid) Rows(rows int32) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Rows(int(rows)))}
}

// Heights sets the optional row heights from top to bottom.
//
//gostrict:host
func (c TGrid) Heights(rowHeights ...Length) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Heights(b4LengthsOf(rowHeights)...))}
}

// Gap sets the space between the rows and the columns equally.
//
//gostrict:host
func (c TGrid) Gap(g Length) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Gap(nui.Length(g)))}
}

// RowGap sets the vertical space between the rows.
//
//gostrict:host
func (c TGrid) RowGap(g Length) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).RowGap(nui.Length(g)))}
}

// ColGap sets the horizontal space between the columns.
//
//gostrict:host
func (c TGrid) ColGap(g Length) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).ColGap(nui.Length(g)))}
}

// Columns sets the amount of columns explicitly.
//
//gostrict:host
func (c TGrid) Columns(cols int32) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Columns(int(cols)))}
}

// Widths sets the optional column widths from left to right.
//
//gostrict:host
func (c TGrid) Widths(colWidths ...Length) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Widths(b4LengthsOf(colWidths)...))}
}

// BackgroundColor sets the background color of the grid.
//
//gostrict:host
func (c TGrid) BackgroundColor(backgroundColor Color) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).BackgroundColor(nui.Color(backgroundColor)).(nui.TGrid))}
}

// Frame sets the size and the position of the grid.
//
//gostrict:host
func (c TGrid) Frame(frame Frame) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Frame(frame.nago()).(nui.TGrid))}
}

// WithFrame changes the size and the position of the grid with fn, which gets the current one.
//
//gostrict:host
func (c TGrid) WithFrame(fn func(Frame) Frame) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TGrid))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TGrid) FullWidth() TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).FullWidth())}
}

// Font sets the font of the texts inside the grid.
//
//gostrict:host
func (c TGrid) Font(font Font) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Font(font.nago()).(nui.TGrid))}
}

// Border sets the border of the grid.
//
//gostrict:host
func (c TGrid) Border(border Border) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Border(border.nago()).(nui.TGrid))}
}

// Visible shows or hides the grid.
//
//gostrict:host
func (c TGrid) Visible(visible bool) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Visible(visible).(nui.TGrid))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TGrid) AccessibilityLabel(label string) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).AccessibilityLabel(label).(nui.TGrid))}
}

// Padding sets the inner spacing around the grid.
//
//gostrict:host
func (c TGrid) Padding(padding Padding) TGrid {
	return TGrid{ref: wrap(get[nui.TGrid](c.ref).Padding(padding.nago()).(nui.TGrid))}
}
