package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
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
	alignment ora.Alignment
}

func GridCell(body core.View) TGridCell {
	return TGridCell{body: body}
}

func (c TGridCell) render(ctx core.RenderContext) ora.GridCell {
	var body ora.Component
	if c.body != nil {
		body = c.body.Render(ctx)
	}

	return ora.GridCell{
		Type:      ora.GridT,
		Body:      body,
		Alignment: c.alignment,
		ColStart:  int64(c.colStart),
		ColEnd:    int64(c.colEnd),
		RowStart:  int64(c.rowStart),
		RowEnd:    int64(c.rowEnd),
		ColSpan:   int64(c.colSpan),
		RowSpan:   int64(c.rowSpan),
	}
}

type TGrid struct {
	cells              []TGridCell
	backgroundColor    ora.Color
	frame              ora.Frame
	rows               int
	cols               int
	gap                ora.Length
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

func (c TGrid) Columns(cols int) TGrid {
	c.cols = cols
	return c
}

func (c TGrid) Render(ctx core.RenderContext) ora.Component {
	cells := make([]ora.GridCell, 0, len(c.cells))
	for _, cell := range c.cells {
		cells = append(cells, cell.render(ctx))
	}

	if c.cols == 0 && c.rows != 0 {
		c.cols = int(math.Round(float64(len(cells)) / float64(c.rows)))
	}

	if c.rows == 0 && c.cols != 0 {
		c.rows = int(math.Round(float64(len(cells)) / float64(c.cols)))
	}

	return ora.Grid{
		Type:    ora.GridT,
		Cells:   cells,
		Rows:    int64(c.rows),
		Columns: int64(c.cols),
		Gap:     c.gap,
	}
}
