package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type TableCell struct {
	id         CID
	body       *Shared[core.Component]
	properties []core.Property
}

func NewTableCell(with func(cell *TableCell)) *TableCell {
	c := &TableCell{
		id: nextPtr(),
	}

	c.body = NewShared[LiveComponent]("body")
	c.properties = []core.Property{c.body}
	if with != nil {
		with(c)
	}

	return c
}

func (c *TableCell) Body() *Shared[core.Component] {
	return c.body
}

func (c *TableCell) ID() CID {
	return c.id
}

func (c *TableCell) Properties(yield func(property core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *TableCell) Render() protocol.Component {
	return c.render()
}
func (c *TableCell) render() protocol.TableCell {
	return protocol.TableCell{
		Ptr:  c.id,
		Type: protocol.TableCellT,
		Body: renderSharedComponent(c.body),
	}
}
