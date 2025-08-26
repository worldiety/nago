// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"math"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// TGridCell is a layout component (Grid Cell).
// It represents a single cell inside a grid layout, defining its
// position, span, alignment, padding, and background color.
type TGridCell struct {
	body            core.View     // the content rendered inside the grid cell
	colStart        int           // starting column (1-based index)
	colEnd          int           // ending column (1-based index)
	rowStart        int           // starting row (1-based index)
	rowEnd          int           // ending row (1-based index)
	colSpan         int           // number of columns to span
	rowSpan         int           // number of rows to span
	padding         proto.Padding // spacing inside the grid cell
	alignment       Alignment     // alignment of the content inside the cell
	backgroundColor Color         // background color of the cell
}

// GridCell creates a new grid cell containing the given body view.
// Rows and columns are 1-based (starting at 1). By default, the cell
// uses Stretch alignment, meaning its content will expand to fill the
// available space. If another alignment is set, the content uses
// wrap-content semantics and the grid background becomes visible.
func GridCell(body core.View) TGridCell {
	return TGridCell{body: body, alignment: Stretch}
}

// Padding sets the inner spacing around the cell content.
func (c TGridCell) Padding(p Padding) TGridCell {
	c.padding = p.ora()
	return c
}

// BackgroundColor sets the background color of the grid cell.
func (c TGridCell) BackgroundColor(color Color) TGridCell {
	c.backgroundColor = color
	return c
}

// Alignment sets the alignment of the content within the grid cell.
func (c TGridCell) Alignment(a Alignment) TGridCell {
	c.alignment = a
	return c
}

// ColStart must start at 1.
func (c TGridCell) ColStart(colStart int) TGridCell {
	c.colStart = colStart
	return c
}

// ColEnd must be always at least +1 of ColStart, even if that column is beyond the defined amount of total columns.
func (c TGridCell) ColEnd(colEnd int) TGridCell {
	c.colEnd = colEnd
	return c
}

// RowStart must start at 1.
func (c TGridCell) RowStart(rowStart int) TGridCell {
	c.rowStart = rowStart
	return c
}

// RowEnd must be always at least +1 of RowStart, even if that row is beyond the defined amount of total rows.
func (c TGridCell) RowEnd(rowEnd int) TGridCell {
	c.rowEnd = rowEnd
	return c
}

// ColSpan behavior is unspecified and can sometime make your life easier, because you must not exactly know
// the layout. However, it may also behave unexpectedly, especially when overlapped.
func (c TGridCell) ColSpan(colSpan int) TGridCell {
	c.colSpan = colSpan
	return c
}

// RowSpan behavior is unspecified and can sometime make your life easier, because you must not exactly know
// the layout. However, it may also behave unexpectedly, especially when overlapped.
func (c TGridCell) RowSpan(rowSpan int) TGridCell {
	c.rowSpan = rowSpan
	return c
}

func (c TGridCell) render(ctx core.RenderContext) proto.GridCell {
	var body proto.Component
	if c.body != nil {
		body = c.body.Render(ctx)
	}

	return proto.GridCell{
		Body:            body,
		ColStart:        proto.Uint(c.colStart),
		ColEnd:          proto.Uint(c.colEnd),
		RowStart:        proto.Uint(c.rowStart),
		RowEnd:          proto.Uint(c.rowEnd),
		ColSpan:         proto.Uint(c.colSpan),
		RowSpan:         proto.Uint(c.rowSpan),
		Padding:         c.padding,
		Alignment:       c.alignment.ora(),
		BackgroundColor: c.backgroundColor.ora(),
	}
}

// TGrid is a layout component (Grid).
// It arranges its children into a grid of rows and columns.
// The grid supports flexible sizing, spacing, alignment, borders,
// accessibility labels, and visibility control. For complex cases
// (e.g., overlapping cells), row/column boundaries must be explicitly defined.
type TGrid struct {
	cells              []TGridCell    // collection of grid cells
	backgroundColor    proto.Color    // background color of the grid
	frame              Frame          // layout frame for size and positioning
	rows               int            // number of rows in the grid
	cols               int            // number of columns in the grid
	colWidths          []proto.Length // optional column widths
	rowGap             proto.Length   // spacing between rows
	colGap             proto.Length   // spacing between columns
	padding            proto.Padding  // inner spacing around the grid
	font               proto.Font     // font applied to text inside the grid
	border             proto.Border   // border styling of the grid
	accessibilityLabel string         // accessibility label for screen readers
	invisible          bool           // when true, the grid is not rendered
}

// Grid creates a new grid containing the given cells.
// For simple cases, only rows or columns need to be set.
// For complex layouts (like overlapping), rows, columns,
// and explicit start/end positions must be defined.
func Grid(cells ...TGridCell) TGrid {
	return TGrid{cells: cells}
}

// Append adds one or more cells to the grid.
func (c TGrid) Append(cells ...TGridCell) TGrid {
	c.cells = append(c.cells, cells...)
	return c
}

// Rows sets the amount of rows explicitly. If not set, the result is undefined. This component tries its
// best to calculate the right amount cells, however, when used with areas, this cannot work properly.
func (c TGrid) Rows(rows int) TGrid {
	c.rows = rows
	return c
}

// Gap sets RowGap and ColGap equally.
func (c TGrid) Gap(g Length) TGrid {
	c.rowGap = g.ora()
	c.colGap = g.ora()
	return c
}

// RowGap sets the vertical spacing between rows in the grid.
func (c TGrid) RowGap(g Length) TGrid {
	c.rowGap = g.ora()
	return c
}

// ColGap sets the horizontal spacing between columns in the grid.
func (c TGrid) ColGap(g Length) TGrid {
	c.colGap = g.ora()
	return c
}

// Columns sets the amount of columns explicitly. If not set, the result is undefined. This component tries its
// best to calculate the right amount cells, however, when used with areas, this cannot work properly.
func (c TGrid) Columns(cols int) TGrid {
	c.cols = cols
	return c
}

// Widths are optional column width from left to right. If not all width are defined, the rest
// of widths are equally distributed based on the remaining amount of space.
func (c TGrid) Widths(colWidths ...Length) TGrid {
	c.colWidths = make([]proto.Length, 0, len(colWidths))
	for _, width := range colWidths {
		c.colWidths = append(c.colWidths, width.ora())
	}
	return c
}

// BackgroundColor sets the background color of the grid.
func (c TGrid) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// Frame sets the layout frame of the grid, including size and positioning.
func (c TGrid) Frame(fr Frame) DecoredView {
	c.frame = fr
	return c
}

// WithFrame applies a transformation function to the grid's frame
// and returns the updated component.
func (c TGrid) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// FullWidth sets the grid to span the full available width.
func (c TGrid) FullWidth() TGrid {
	c.frame.Width = "100%"
	return c
}

// Font sets the font style applied to text inside the grid.
func (c TGrid) Font(font Font) DecoredView {
	c.font = font.ora()
	return c
}

// Border sets the border styling of the grid.
func (c TGrid) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

// Visible controls the visibility of the grid; setting false hides it.
func (c TGrid) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// AccessibilityLabel sets the accessibility label of the grid,
// used by screen readers.
func (c TGrid) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// Padding sets the inner spacing around the grid.
func (c TGrid) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

// cellCount calculates the number of grid cells based on spans and column positions.
// It accounts for cells with explicit colSpan or colStart/colEnd values, falling back
// to counting as one cell if no span information is provided.
func (c TGrid) cellCount() int {
	count := 0
	for _, cell := range c.cells {
		if cell.colSpan != 0 {
			count += cell.colSpan
			continue
		}

		if delta := cell.colEnd - cell.colStart; delta > 0 {
			count += delta
			continue
		}

		count++
	}

	return count
}

// Render builds and returns the protocol representation of the grid.
func (c TGrid) Render(ctx core.RenderContext) core.RenderNode {
	cells := make([]proto.GridCell, 0, len(c.cells))
	for _, cell := range c.cells {
		cells = append(cells, cell.render(ctx))
	}

	virtualCellCount := c.cellCount()
	if c.cols == 0 && c.rows != 0 {
		c.cols = int(math.Round(float64(virtualCellCount) / float64(c.rows)))
	}

	if c.rows == 0 && c.cols != 0 {
		c.rows = int(math.Round(float64(virtualCellCount) / float64(c.cols)))
	}

	if c.rows < 0 {
		c.rows = 0
	}

	if c.cols < 0 {
		c.cols = 0
	}

	return &proto.Grid{
		Cells:              cells,
		Rows:               proto.Uint(c.rows),
		Columns:            proto.Uint(c.cols),
		RowGap:             c.rowGap,
		ColGap:             c.colGap,
		Frame:              c.frame.ora(),
		BackgroundColor:    c.backgroundColor,
		Padding:            c.padding,
		Border:             c.border,
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Invisible:          proto.Bool(c.invisible),
		Font:               c.font,
		ColWidths:          c.colWidths,
	}
}
