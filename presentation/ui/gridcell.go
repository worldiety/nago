package ui

import "go.wdy.de/nago/container/slice"

type GridCell struct {
	id         CID
	body       *Shared[LiveComponent]
	colStart   Int
	colEnd     Int
	rowStart   Int
	rowEnd     Int
	properties slice.Slice[Property]
	functions  slice.Slice[*Func]
}

func NewGridCell(with func(cell *GridCell)) *GridCell {
	c := &GridCell{
		id: nextPtr(),
	}

	c.body = NewShared[LiveComponent]("body")
	c.colStart = NewShared[int64]("colStart")
	c.colEnd = NewShared[int64]("colEnd")
	c.rowStart = NewShared[int64]("rowStart")
	c.rowEnd = NewShared[int64]("rowEnd")
	c.properties = slice.Of[Property](c.body, c.colStart, c.colEnd, c.rowStart, c.rowEnd)
	c.functions = slice.Of[*Func]()
	if with != nil {
		with(c)
	}

	return c
}

func (c *GridCell) Body() *Shared[LiveComponent] {
	return c.body
}

func (c *GridCell) ColStart() Int {
	return c.colStart
}

func (c *GridCell) ColEnd() Int {
	return c.colEnd
}

func (c *GridCell) RowStart() Int {
	return c.rowStart
}

func (c *GridCell) RowEnd() Int {
	return c.rowEnd
}

func (c *GridCell) ID() CID {
	return c.id
}

func (c *GridCell) Type() string {
	return "GridCell"
}

func (c *GridCell) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *GridCell) Children() slice.Slice[LiveComponent] {
	return slice.Of(c.body.v)
}

func (c *GridCell) Functions() slice.Slice[*Func] {
	return c.functions
}
