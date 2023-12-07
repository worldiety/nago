package ui

import "go.wdy.de/nago/container/slice"

type TableRow struct {
	id         CID
	cells      *SharedList[*TableCell]
	properties slice.Slice[Property]
}

func NewTableRow(with func(row *TableRow)) *TableRow {
	c := &TableRow{
		id: nextPtr(),
	}

	c.cells = NewSharedList[*TableCell]("cells")
	c.properties = slice.Of[Property](c.cells)
	if with != nil {
		with(c)
	}

	return c
}

func (c *TableRow) Cells() *SharedList[*TableCell] {
	return c.cells
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
