package ui

import "go.wdy.de/nago/container/slice"

type Grid struct {
	id         CID
	cells      *SharedList[LiveComponent]
	rows       Int
	columns    Int
	gap        *Shared[Size]
	properties slice.Slice[Property]
	functions  slice.Slice[*Func]
}

func NewGrid(with func(grid *Grid)) *Grid {
	c := &Grid{
		id: nextPtr(),
	}

	c.cells = NewSharedList[LiveComponent]("cells")
	c.rows = NewShared[int64]("rows")
	c.columns = NewShared[int64]("columns")
	c.gap = NewShared[Size]("gap")
	c.properties = slice.Of[Property](c.cells, c.rows, c.columns, c.gap)
	c.functions = slice.Of[*Func]()
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

func (c *Grid) ID() CID {
	return c.id
}

func (c *Grid) Type() string {
	return "Grid"
}

func (c *Grid) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *Grid) Children() slice.Slice[LiveComponent] {
	return slice.Of(c.cells.values...)
}

func (c *Grid) Functions() slice.Slice[*Func] {
	return c.functions
}
