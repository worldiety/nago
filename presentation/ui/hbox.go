package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type HBox struct {
	id         CID
	children   *SharedList[core.Component]
	alignment  String
	properties []core.Property
}

func NewHBox(with func(hbox *HBox)) *HBox {
	c := &HBox{
		id:        nextPtr(),
		children:  NewSharedList[core.Component]("children"),
		alignment: NewShared[string]("alignment"),
	}
	c.alignment.Set("grid")
	c.alignment.SetDirty(false)

	c.properties = []core.Property{c.children, c.alignment}
	if with != nil {
		with(c)
	}

	return c
}

func (c *HBox) Append(children ...LiveComponent) *HBox {
	c.children.Append(children...)
	return c
}

func (c *HBox) Children() *SharedList[LiveComponent] {
	return c.children
}

// Alignment is a layout hint. Supported values are "grid" (default) or "flex-left", "flex-right" or "flex-center".
func (c *HBox) Alignment() String {
	return c.alignment
}

func (c *HBox) ID() CID {
	return c.id
}

func (c *HBox) Type() string {
	return "HBox"
}

func (c *HBox) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *HBox) Render() protocol.Component {
	panic("not implemented")
}
