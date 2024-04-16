package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type VBox struct {
	id         CID
	children   *SharedList[core.Component]
	properties []core.Property
}

func NewVBox(with func(vbox *VBox)) *VBox {
	c := &VBox{
		id:       nextPtr(),
		children: NewSharedList[core.Component]("children"),
	}

	c.properties = []core.Property{c.children}
	if with != nil {
		with(c)
	}
	return c
}

func (c *VBox) Append(children ...core.Component) *VBox {
	c.children.Append(children...)
	return c
}

func (c *VBox) Children() *SharedList[core.Component] {
	return c.children
}

func (c *VBox) ID() CID {
	return c.id
}

func (c *VBox) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *VBox) Render() ora.Component {
	return ora.VBox{
		Ptr:      c.id,
		Type:     ora.VBoxT,
		Children: renderSharedListComponents(c.children),
	}
}
