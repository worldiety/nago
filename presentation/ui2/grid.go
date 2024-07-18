package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"math"
)

type TGridCell struct {
	body     core.View
	colStart int
	colEnd   int
	rowStart int
	rowEnd   int
	colSpan  int
	rowSpan  int
	padding  ora.Padding
}

func GridCell(body core.View) TGridCell {
	return TGridCell{body: body}
}

func (c TGridCell) ColStart(colStart int) TGridCell {
	c.colStart = colStart
	return c
}
func (c TGridCell) Padding(p ora.Padding) TGridCell {
	c.padding = p
	return c
}

func (c TGridCell) ColEnd(colEnd int) TGridCell {
	c.colEnd = colEnd
	return c
}

func (c TGridCell) RowStart(rowStart int) TGridCell {
	c.rowStart = rowStart
	return c
}

func (c TGridCell) RowEnd(rowEnd int) TGridCell {
	c.rowEnd = rowEnd
	return c
}

func (c TGridCell) ColSpan(colSpan int) TGridCell {
	c.colSpan = colSpan
	return c
}

func (c TGridCell) RowSpan(rowSpan int) TGridCell {
	c.rowSpan = rowSpan
	return c
}

func (c TGridCell) render(ctx core.RenderContext) ora.GridCell {
	var body ora.Component
	if c.body != nil {
		body = c.body.Render(ctx)
	}

	return ora.GridCell{
		Type:     ora.GridT,
		Body:     body,
		ColStart: int64(c.colStart),
		ColEnd:   int64(c.colEnd),
		RowStart: int64(c.rowStart),
		RowEnd:   int64(c.rowEnd),
		ColSpan:  int64(c.colSpan),
		RowSpan:  int64(c.rowSpan),
		Padding:  c.padding,
	}
}

type TGrid struct {
	cells              []TGridCell
	backgroundColor    ora.Color
	frame              ora.Frame
	rows               int
	cols               int
	colWidths          []ora.Length
	rowGap             ora.Length
	colGap             ora.Length
	padding            ora.Padding
	font               ora.Font
	border             ora.Border
	accessibilityLabel string
	invisible          bool
}

func Grid(cells ...TGridCell) TGrid {
	return TGrid{cells: cells}
}

func (c TGrid) Rows(rows int) TGrid {
	c.rows = rows
	return c
}

func (c TGrid) Gap(g ora.Length) TGrid {
	c.rowGap = g
	c.colGap = g
	return c
}

func (c TGrid) RowGap(g ora.Length) TGrid {
	c.rowGap = g
	return c
}

func (c TGrid) ColGap(g ora.Length) TGrid {
	c.colGap = g
	return c
}

func (c TGrid) Columns(cols int) TGrid {
	c.cols = cols
	return c
}

func (c TGrid) Widths(colWidths ...ora.Length) TGrid {
	c.colWidths = colWidths
	return c
}

func (c TGrid) BackgroundColor(backgroundColor ora.Color) core.DecoredView {
	c.backgroundColor = backgroundColor
	return c
}

func (c TGrid) Frame(fr ora.Frame) core.DecoredView {
	c.frame = fr
	return c
}

func (c TGrid) Font(font ora.Font) core.DecoredView {
	c.font = font
	return c
}

func (c TGrid) Border(border ora.Border) core.DecoredView {
	c.border = border
	return c
}

func (c TGrid) Visible(visible bool) core.DecoredView {
	c.invisible = !visible
	return c
}

func (c TGrid) AccessibilityLabel(label string) core.DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TGrid) Padding(padding ora.Padding) core.DecoredView {
	c.padding = padding
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

func (c TGrid) Render(ctx core.RenderContext) ora.Component {
	cells := make([]ora.GridCell, 0, len(c.cells))
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

	return ora.Grid{
		Type:               ora.GridT,
		Cells:              cells,
		Rows:               int64(c.rows),
		Columns:            int64(c.cols),
		RowGap:             c.rowGap,
		ColGap:             c.colGap,
		Frame:              c.frame,
		BackgroundColor:    c.backgroundColor,
		Padding:            c.padding,
		Border:             c.border,
		AccessibilityLabel: c.accessibilityLabel,
		Invisible:          c.invisible,
		Font:               c.font,
		ColWidths:          c.colWidths,
	}
}
