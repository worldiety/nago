package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Grid struct {
	id         ora.Ptr
	cells      *SharedList[*GridCell]
	rows       Int
	columns    Int
	smColumns  Int
	mdColumns  Int
	lgColumns  Int
	gap        *Shared[Size]
	properties []core.Property
}

func NewGrid(with func(grid *Grid)) *Grid {
	c := &Grid{
		id: nextPtr(),
	}

	c.cells = NewSharedList[*GridCell]("cells")
	c.rows = NewShared[int64]("rows")
	c.columns = NewShared[int64]("columns")
	c.smColumns = NewShared[int64]("smColumns")
	c.mdColumns = NewShared[int64]("mdColumns")
	c.lgColumns = NewShared[int64]("lgColumns")
	c.gap = NewShared[Size]("gap")
	c.properties = []core.Property{c.cells, c.rows, c.columns, c.gap, c.smColumns, c.mdColumns, c.lgColumns}
	if with != nil {
		with(c)
	}

	return c
}

func (c *Grid) AppendCell(cell *GridCell) *Grid {
	c.cells.Append(cell)
	return c
}

func (c *Grid) AppendCells(cells ...*GridCell) *Grid {
	for _, cell := range cells {
		c.cells.Append(cell)
	}
	return c
}

func (c *Grid) Rows() Int {
	return c.rows
}

func (c *Grid) Columns() Int {
	return c.columns
}

func (c *Grid) ColumnsSmallOrLarger() Int {
	return c.smColumns
}

func (c *Grid) ColumnsMediumOrLarger() Int {
	return c.mdColumns
}

func (c *Grid) ColumnsLarger() Int {
	return c.lgColumns
}

func (c *Grid) ID() ora.Ptr {
	return c.id
}

func (c *Grid) Type() ora.ComponentType {
	return ora.GridT
}

func (c *Grid) Properties(yield func(property core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Grid) Render() ora.Component {
	return c.render()
}

func (c *Grid) render() ora.Grid {
	var cells []ora.GridCell
	c.cells.Iter(func(component *GridCell) bool {
		cells = append(cells, component.render())
		return true
	})

	return ora.Grid{
		Ptr:  c.id,
		Type: ora.GridT,
		Cells: ora.Property[[]ora.GridCell]{
			Ptr:   c.cells.ID(),
			Value: cells,
		},
		Rows:      c.rows.render(),
		Columns:   c.columns.render(),
		SMColumns: c.smColumns.render(),
		MDColumns: c.mdColumns.render(),
		LGColumns: c.lgColumns.render(),
		Gap: ora.Property[string]{
			Ptr:   c.gap.ID(),
			Value: string(c.gap.v),
		},
	}
}
