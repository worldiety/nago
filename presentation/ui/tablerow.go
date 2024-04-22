package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TableRow struct {
	id         ora.Ptr
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

func (c *TableRow) ID() ora.Ptr {
	return c.id
}

func (c *TableRow) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *TableRow) Render() ora.Component {
	return c.render()
}

func (c *TableRow) render() ora.TableRow {
	var cells []ora.TableCell
	c.cells.Iter(func(cell *TableCell) bool {
		cells = append(cells, cell.render())
		return true
	})
	return ora.TableRow{
		Ptr:  c.id,
		Type: ora.TableRowT,
		Cells: ora.Property[[]ora.TableCell]{
			Ptr:   c.cells.ID(),
			Value: cells,
		},
	}
}
