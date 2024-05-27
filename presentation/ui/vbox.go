package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

// deprecated use FlexContainer
type VBox struct {
	id         ora.Ptr
	children   *SharedList[core.Component]
	properties []core.Property
}

// deprecated use FlexContainer or NewVStack()
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

func (c *VBox) ID() ora.Ptr {
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
