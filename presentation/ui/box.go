package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"slices"
)

type AlignedComponent struct {
	Component core.View
	Alignment Alignment
}

type Box struct {
	id              ora.Ptr
	children        []AlignedComponent
	properties      []core.Property
	backgroundColor ora.Color
	frame           ora.Frame
	Padding         ora.Padding
}

func NewBox(with func(box *Box)) *Box {
	c := &Box{
		id: nextPtr(),
	}

	c.properties = []core.Property{}
	if with != nil {
		with(c)
	}

	return c
}

func (c *Box) BackgroundColor() ora.Color {
	return c.backgroundColor
}

func (c *Box) SetBackgroundColor(backgroundColor ora.Color) {
	c.backgroundColor = backgroundColor
}

// Align adds the given child with the defined alignment. Added Order is Z-Order.
// Aligning multiple children is not allowed, thus any other occurrence is removed and the new child
// is appended.
func (c *Box) Align(alignment Alignment, child core.View) {
	slices.DeleteFunc(c.children, func(component AlignedComponent) bool {
		return component.Alignment == alignment
	})

	c.children = append(c.children, AlignedComponent{
		Component: child,
		Alignment: alignment,
	})
}

func (c *Box) ID() ora.Ptr {
	return c.id
}

func (c *Box) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Box) Frame() *ora.Frame {
	return &c.frame
}

func (c *Box) Render() ora.Component {
	var tmp []ora.AlignedComponent
	for _, child := range c.children {
		tmp = append(tmp, ora.AlignedComponent{
			Component: child.Component.Render(),
			Alignment: child.Alignment,
		})
	}

	return ora.Box{
		Type:            ora.BoxT,
		Children:        tmp,
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Padding:         c.Padding,
	}
}
