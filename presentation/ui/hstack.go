package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type HStack struct {
	id              ora.Ptr
	children        *SharedList[core.Component] // TODO why is this shared? do we need the dirty flag? do we need a pointer box? why not always render dirty?
	properties      []core.Property
	alignment       ora.Alignment
	backgroundColor ora.NamedColor
	frame           ora.Frame
	gap             ora.Length
	padding         ora.Padding
}

func NewHStack(with func(hstack *HStack)) *HStack {
	c := &HStack{
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

func (c *HStack) Padding() ora.Padding {
	return c.padding
}

func (c *HStack) SetPadding(padding ora.Padding) {
	c.padding = padding
}

func (c *HStack) Gap() ora.Length {
	return c.gap
}

func (c *HStack) SetGap(gap ora.Length) {
	c.gap = gap
}

func (c *HStack) BackgroundColor() ora.NamedColor {
	return c.backgroundColor
}

func (c *HStack) SetBackgroundColor(backgroundColor ora.NamedColor) {
	c.backgroundColor = backgroundColor
}

func (c *HStack) Alignment() ora.Alignment {
	return c.alignment
}

func (c *HStack) SetAlignment(alignment ora.Alignment) {
	c.alignment = alignment
}

func (c *HStack) Append(children ...core.Component) {
	// this signature does not return builder pattern anymore, because it makes polymorphic interface usage impossible
	c.children.Append(children...)
}

func (c *HStack) Children() *SharedList[core.Component] {
	return c.children
}

func (c *HStack) ID() ora.Ptr {
	return c.id
}

func (c *HStack) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *HStack) Frame() *ora.Frame {
	return &c.frame
}

func (c *HStack) Render() ora.Component {
	return ora.HStack{
		Type:            ora.HStackT,
		Children:        renderSharedListComponentsFlat(c.children),
		Frame:           c.frame,
		Alignment:       c.alignment,
		BackgroundColor: c.backgroundColor,
		Gap:             c.gap,
		Padding:         c.padding,
	}
}
