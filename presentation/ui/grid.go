package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"math"
)

type TGridCell struct {
	body      core.View
	colStart  int
	colEnd    int
	rowStart  int
	rowEnd    int
	colSpan   int
	rowSpan   int
	padding   proto.Padding
	alignment Alignment
}

// GridCell creates a cell based on the given body. Rows and Columns start at 1, not zero.
// Without any alignment rules, the cell will stretch its body automatically to the calculated
// cell dimensions. Otherwise, if a cell alignment is set, the size is wrap-content semantics
// and the background of the grid will be visible. Thus, the default specification of no-alignment
// is different here.
func GridCell(body core.View) TGridCell {
	return TGridCell{body: body, alignment: Stretch}
}

func (c TGridCell) Padding(p Padding) TGridCell {
	c.padding = p.ora()
	return c
}

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
		Body:      body,
		ColStart:  proto.Uint(c.colStart),
		ColEnd:    proto.Uint(c.colEnd),
		RowStart:  proto.Uint(c.rowStart),
		RowEnd:    proto.Uint(c.rowEnd),
		ColSpan:   proto.Uint(c.colSpan),
		RowSpan:   proto.Uint(c.rowSpan),
		Padding:   c.padding,
		Alignment: c.alignment.ora(),
	}
}

type TGrid struct {
	cells              []TGridCell
	backgroundColor    proto.Color
	frame              proto.Frame
	rows               int
	cols               int
	colWidths          []proto.Length
	rowGap             proto.Length
	colGap             proto.Length
	padding            proto.Padding
	font               proto.Font
	border             proto.Border
	accessibilityLabel string
	invisible          bool
}

// Grid is a complex component and hard to control. For simple situations, it usually works by just setting the
// required amount of rows or columns. However, in complex (e.g. overlapping) situations, you must be as explicit
// as possible and define the exact amount of rows and columns and for each cell the row-start/end and col-start/end.
func Grid(cells ...TGridCell) TGrid {
	return TGrid{cells: cells}
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

func (c TGrid) RowGap(g Length) TGrid {
	c.rowGap = g.ora()
	return c
}

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

func (c TGrid) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TGrid) Frame(fr Frame) DecoredView {
	c.frame = fr.ora()
	return c
}

func (c TGrid) FullWidth() TGrid {
	c.frame.Width = "100%"
	return c
}

func (c TGrid) Font(font Font) DecoredView {
	c.font = font.ora()
	return c
}

func (c TGrid) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c TGrid) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TGrid) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TGrid) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

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

	return &proto.Grid{
		Cells:              cells,
		Rows:               proto.Uint(c.rows),
		Columns:            proto.Uint(c.cols),
		RowGap:             c.rowGap,
		ColGap:             c.colGap,
		Frame:              c.frame,
		BackgroundColor:    c.backgroundColor,
		Padding:            c.padding,
		Border:             c.border,
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Invisible:          proto.Bool(c.invisible),
		Font:               c.font,
		ColWidths:          c.colWidths,
	}
}
