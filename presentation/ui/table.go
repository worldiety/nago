package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type Table struct {
	id         CID
	headers    *SharedList[*TableCell]
	rows       *SharedList[*TableRow]
	properties []core.Property
}

func NewTable(with func(table *Table)) *Table {
	c := &Table{
		id: nextPtr(),
	}

	c.rows = NewSharedList[*TableRow]("rows")
	c.headers = NewSharedList[*TableCell]("headers")
	c.properties = []core.Property{c.headers, c.rows}
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

func (c *Table) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Table) Render() protocol.Component {
	panic("not implemented")
}
