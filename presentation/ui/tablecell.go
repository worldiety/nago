package ui

import "go.wdy.de/nago/container/slice"

type TableCell struct {
	id         CID
	body       *Shared[LiveComponent]
	properties slice.Slice[Property]
}

func NewTableCell(with func(cell *TableCell)) *TableCell {
	c := &TableCell{
		id: nextPtr(),
	}

	c.body = NewShared[LiveComponent]("body")
	c.properties = slice.Of[Property](c.body)
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
