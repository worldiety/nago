package ui

import "go.wdy.de/nago/container/slice"

type TableRow struct {
	id         CID
	cells      *SharedList[LiveComponent]
	properties slice.Slice[Property]
	functions  slice.Slice[*Func]
}

func NewTableRow(with func(row *TableRow)) *TableRow {
	c := &TableRow{
		id: nextPtr(),
	}

	c.cells = NewSharedList[LiveComponent]("cells")
	c.properties = slice.Of[Property](c.cells)
	c.functions = slice.Of[*Func]()
	if with != nil {
		with(c)
	}

	return c
}

func (c *TableRow) AppendCell(cell *TableCell) *TableRow {
	c.cells.Append(cell)
	return c
}

func (c *TableRow) ID() CID {
	return c.id
}

func (c *TableRow) Type() string {
	return "TableRow"
}

func (c *TableRow) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *TableRow) Children() slice.Slice[LiveComponent] {
	return slice.Of(c.cells.values...)
}

func (c *TableRow) Functions() slice.Slice[*Func] {
	return c.functions
}
