package uilegacy

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TableCell struct {
	id         ora.Ptr
	body       *Shared[core.View]
	properties []core.Property
}

// NewTextCell is a handy shortcut to just put some default text into a cell.
func NewTextCell(text string) *TableCell {
	return NewTableCell(func(cell *TableCell) {
		cell.Body().Set(NewStr(text))
	})
}

func NewTableCell(with func(cell *TableCell)) *TableCell {
	c := &TableCell{
		id: nextPtr(),
	}

	c.body = NewShared[core.View]("body")
	c.properties = []core.Property{c.body}
	if with != nil {
		with(c)
	}

	return c
}

func (c *TableCell) Body() *Shared[core.View] {
	return c.body
}

func (c *TableCell) ID() ora.Ptr {
	return c.id
}

func (c *TableCell) Properties(yield func(property core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *TableCell) Render() ora.Component {
	return c.render()
}
func (c *TableCell) render() ora.TableCell {
	return ora.TableCell{
		Ptr:  c.id,
		Type: ora.TableCellT,
		Body: renderSharedComponent(c.body),
	}
}
