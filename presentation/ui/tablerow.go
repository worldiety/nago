package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type TableRow struct {
	id         CID
	cells      *SharedList[*TableCell]
	properties []core.Property
}

func NewTableRow(with func(row *TableRow)) *TableRow {
	c := &TableRow{
		id: nextPtr(),
	}

	c.cells = NewSharedList[*TableCell]("cells")
	c.properties = []core.Property{c.cells}
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

func (c *TableRow) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *TableRow) Render() protocol.Component {
	return c.render()
}

func (c *TableRow) render() protocol.TableRow {
	var cells []protocol.TableCell
	c.cells.Iter(func(cell *TableCell) bool {
		cells = append(cells, cell.render())
		return true
	})
	return protocol.TableRow{
		Ptr:  c.id,
		Type: protocol.TableRowT,
		Cells: protocol.Property[[]protocol.TableCell]{
			Ptr:   c.cells.ID(),
			Value: cells,
		},
	}
}
