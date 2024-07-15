package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type GridCell struct {
	id         ora.Ptr
	body       *Shared[core.View]
	colStart   Int
	colEnd     Int
	colSpan    Int
	smColSpan  Int
	mdColSpan  Int
	lgColSpan  Int
	rowStart   Int
	rowEnd     Int
	properties []core.Property
}

func NewGridCell(with func(cell *GridCell)) *GridCell {
	c := &GridCell{
		id: nextPtr(),
	}

	c.body = NewShared[core.View]("body")
	c.colStart = NewShared[int64]("colStart")
	c.colEnd = NewShared[int64]("colEnd")
	c.rowStart = NewShared[int64]("rowStart")
	c.rowEnd = NewShared[int64]("rowEnd")
	c.colSpan = NewShared[int64]("colSpan")
	c.smColSpan = NewShared[int64]("smColSpan")
	c.mdColSpan = NewShared[int64]("mdColSpan")
	c.lgColSpan = NewShared[int64]("lgColSpan")
	c.properties = []core.Property{c.body, c.colStart, c.colEnd, c.rowStart, c.rowEnd, c.colSpan, c.smColSpan, c.mdColSpan, c.lgColSpan}
	if with != nil {
		with(c)
	}

	return c
}

func (c *GridCell) Body() *Shared[core.View] {
	return c.body
}

func (c *GridCell) ColStart() Int {
	return c.colStart
}

func (c *GridCell) ColEnd() Int {
	return c.colEnd
}

// ColSpan for any device size.
func (c *GridCell) ColSpan() Int {
	return c.colSpan
}

// SmallColSpan is the span for any device equal or larger small size.
func (c *GridCell) SmallColSpan() Int {
	return c.smColSpan
}

// MediumColSpan is the span for any device equal or larger medium size.
func (c *GridCell) MediumColSpan() Int {
	return c.mdColSpan
}

// LargeColSpan is the span for any device equal or larger large size.
func (c *GridCell) LargeColSpan() Int {
	return c.lgColSpan
}

func (c *GridCell) RowStart() Int {
	return c.rowStart
}

func (c *GridCell) RowEnd() Int {
	return c.rowEnd
}

func (c *GridCell) ID() ora.Ptr {
	return c.id
}

func (c *GridCell) Type() ora.ComponentType {
	return ora.GridCellT
}

func (c *GridCell) Properties(yield func(property core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *GridCell) Render() ora.Component {
	return c.render()
}

func (c *GridCell) render() ora.GridCell {
	return ora.GridCell{
		Ptr:       c.id,
		Type:      ora.GridCellT,
		Body:      renderSharedComponent(c.body),
		ColStart:  c.colStart.render(),
		ColEnd:    c.colEnd.render(),
		RowStart:  c.rowStart.render(),
		RowEnd:    c.rowEnd.render(),
		ColSpan:   c.colSpan.render(),
		SmColSpan: c.smColSpan.render(),
		MdColSpan: c.mdColSpan.render(),
		LgColSpan: c.lgColSpan.render(),
	}
}
