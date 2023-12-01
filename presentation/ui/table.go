package ui

import "go.wdy.de/nago/container/slice"

type Table struct {
	id         CID
	headers    *SharedList[LiveComponent]
	rows       *SharedList[LiveComponent]
	properties slice.Slice[Property]
	functions  slice.Slice[*Func]
}

func NewTable(with func(table *Table)) *Table {
	c := &Table{
		id: nextPtr(),
	}

	c.rows = NewSharedList[LiveComponent]("rows")
	c.headers = NewSharedList[LiveComponent]("headers")
	c.properties = slice.Of[Property](c.headers, c.rows)
	c.functions = slice.Of[*Func]()
	if with != nil {
		with(c)
	}

	return c
}

func (c *Table) AppendColumn(cell *TableCell) *Table {
	c.headers.Append(cell)
	return c
}

func (c *Table) AppendColumns(cells ...*TableCell) *Table {
	for _, cell := range cells {
		c.headers.Append(cell)
	}
	return c
}

func (c *Table) AppendRow(row *TableRow) *Table {
	c.rows.Append(row)
	return c
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

func (c *Table) Children() slice.Slice[LiveComponent] {
	tmp := make([]LiveComponent, 0, c.rows.Len()+c.headers.Len())
	tmp = append(tmp, c.rows.values...)
	tmp = append(tmp, c.headers.values...)
	return slice.Of(tmp...)
}

func (c *Table) Functions() slice.Slice[*Func] {
	return c.functions
}
