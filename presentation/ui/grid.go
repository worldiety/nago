package ui

import "go.wdy.de/nago/container/slice"

type Grid struct {
	id         CID
	cells      *SharedList[LiveComponent]
	rows       Int
	columns    Int
	smColumns  Int
	mdColumns  Int
	lgColumns  Int
	gap        *Shared[Size]
	properties slice.Slice[Property]
}

func NewGrid(with func(grid *Grid)) *Grid {
	c := &Grid{
		id: nextPtr(),
	}

	c.cells = NewSharedList[LiveComponent]("cells")
	c.rows = NewShared[int64]("rows")
	c.columns = NewShared[int64]("columns")
	c.smColumns = NewShared[int64]("smColumns")
	c.mdColumns = NewShared[int64]("mdColumns")
	c.lgColumns = NewShared[int64]("lgColumns")
	c.gap = NewShared[Size]("gap")
	c.properties = slice.Of[Property](c.cells, c.rows, c.columns, c.gap, c.smColumns, c.mdColumns, c.lgColumns)
	if with != nil {
		with(c)
	}

	return c
}

func (c *Grid) AppendCell(cell *GridCell) *Grid {
	c.cells.Append(cell)
	return c
}

func (c *Grid) AppendCells(cells ...*GridCell) *Grid {
	for _, cell := range cells {
		c.cells.Append(cell)
	}
	return c
}

func (c *Grid) Rows() Int {
	return c.rows
}

func (c *Grid) Columns() Int {
	return c.columns
}

func (c *Grid) ColumnsSmallOrLarger() Int {
	return c.smColumns
}

func (c *Grid) ColumnsMediumOrLarger() Int {
	return c.mdColumns
}

func (c *Grid) ColumnsLarger() Int {
	return c.lgColumns
}

func (c *Grid) ID() CID {
	return c.id
}

func (c *Grid) Type() string {
	return "Grid"
}

func (c *Grid) Properties() slice.Slice[Property] {
	return c.properties
}
