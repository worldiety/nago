package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

// deprecated: use FlexContainer
type HBox struct {
	id         ora.Ptr
	children   *SharedList[core.Component]
	alignment  String
	properties []core.Property
}

// deprecated: use FlexContainer or NewHStack()
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

func (c *HBox) Append(children ...core.Component) *HBox {
	c.children.Append(children...)
	return c
}

func (c *HBox) Children() *SharedList[core.Component] {
	return c.children
}

// Alignment is a layout hint. Supported values are "grid" (default) or "flex-left", "flex-right" or "flex-center".
func (c *HBox) Alignment() String {
	return c.alignment
}

func (c *HBox) ID() ora.Ptr {
	return c.id
}

func (c *HBox) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *HBox) Render() ora.Component {
	return ora.HBox{
		Ptr:       c.id,
		Type:      ora.HBoxT,
		Children:  renderSharedListComponents(c.children),
		Alignment: c.alignment.render(),
	}
}
