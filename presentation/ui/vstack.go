package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type VStack struct {
	id              ora.Ptr
	children        *SharedList[core.Component] // TODO why is this shared? do we need the dirty flag? do we need a pointer box? why not always render dirty?
	properties      []core.Property
	alignment       ora.Alignment
	backgroundColor ora.NamedColor
	frame           ora.Frame
}

func NewVStack(with func(hstack *VStack)) *VStack {
	c := &VStack{
		id:       nextPtr(),
		children: NewSharedList[core.Component]("children"),
	}

	c.alignment = "" // if nothing is defined, ora.Center must be applied by renderer
	c.properties = []core.Property{c.children}
	if with != nil {
		with(c)
	}

	return c
}

func (c *VStack) BackgroundColor() ora.NamedColor {
	return c.backgroundColor
}

func (c *VStack) SetBackgroundColor(backgroundColor ora.NamedColor) {
	c.backgroundColor = backgroundColor
}

func (c *VStack) Alignment() ora.Alignment {
	return c.alignment
}

func (c *VStack) SetAlignment(alignment ora.Alignment) {
	c.alignment = alignment
}

func (c *VStack) Append(children ...core.Component) {
	// this signature does not return builder pattern anymore, because it makes polymorphic interface usage impossible
	c.children.Append(children...)
}

func (c *VStack) Children() *SharedList[core.Component] {
	return c.children
}

func (c *VStack) ID() ora.Ptr {
	return c.id
}

func (c *VStack) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *VStack) Frame() *ora.Frame {
	return &c.frame
}

func (c *VStack) Render() ora.Component {
	return ora.HStack{
		Ptr:             c.id,
		Type:            ora.VStackT,
		Children:        renderSharedListComponentsFlat(c.children),
		Frame:           c.frame,
		Alignment:       c.alignment,
		BackgroundColor: c.backgroundColor,
	}
}
