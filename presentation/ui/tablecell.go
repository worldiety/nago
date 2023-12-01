package ui

import "go.wdy.de/nago/container/slice"

type TableCell struct {
	id         CID
	body       *Shared[LiveComponent]
	properties slice.Slice[Property]
	functions  slice.Slice[*Func]
}

func NewTableCell(with func(cell *TableCell)) *TableCell {
	c := &TableCell{
		id: nextPtr(),
	}

	c.body = NewShared[LiveComponent]("body")
	c.properties = slice.Of[Property](c.body)
	c.functions = slice.Of[*Func]()
	if with != nil {
		with(c)
	}

	return c
}

func (c *TableCell) Body() *Shared[LiveComponent] {
	return c.body
}

func (c *TableCell) ID() CID {
	return c.id
}

func (c *TableCell) Type() string {
	return "TableCell"
}

func (c *TableCell) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *TableCell) Children() slice.Slice[LiveComponent] {
	return slice.Of(c.body.v)
}

func (c *TableCell) Functions() slice.Slice[*Func] {
	return c.functions
}
