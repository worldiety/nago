package ui

import "go.wdy.de/nago/container/slice"

type Table struct {
	id         CID
	headers    *SharedList[*TableCell]
	rows       *SharedList[*TableRow]
	properties slice.Slice[Property]
}

func NewTable(with func(table *Table)) *Table {
	c := &Table{
		id: nextPtr(),
	}

	c.rows = NewSharedList[*TableRow]("rows")
	c.headers = NewSharedList[*TableCell]("headers")
	c.properties = slice.Of[Property](c.headers, c.rows)
	if with != nil {
		with(c)
	}

	return c
}

func (c *Table) Rows() *SharedList[*TableRow] {
	return c.rows
}

func (c *Table) Header() *SharedList[*TableCell] {
	return c.headers
}

func (c *Table) ID() CID {
	return c.id
}

func (c *Table) Type() string {
	return "Table"
}

func (c *Table) Properties() slice.Slice[Property] {
	return c.properties
}
