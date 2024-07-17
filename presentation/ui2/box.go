package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"slices"
)

type AlignedComponent struct {
	Component core.View
	Alignment ora.Alignment
}

type Box struct {
	children        []AlignedComponent
	backgroundColor ora.Color
	frame           ora.Frame
	padding         ora.Padding
	with            func(box *Box)
}

func NewBox(with func(box *Box)) *Box {
	c := &Box{}

	return c
}

func (c *Box) BackgroundColor(backgroundColor ora.Color) {
	c.backgroundColor = backgroundColor
}

// Align adds the given child with the defined alignment. Added Order is Z-Order.
// Aligning multiple children is not allowed, thus any other occurrence is removed and the new child
// is appended.
func (c *Box) Align(alignment ora.Alignment, child core.View) {
	slices.DeleteFunc(c.children, func(component AlignedComponent) bool {
		return component.Alignment == alignment
	})

	c.children = append(c.children, AlignedComponent{
		Component: child,
		Alignment: alignment,
	})
}

func (c *Box) Frame(f ora.Frame) {
	c.frame = f
}

func (c *Box) Padding(p ora.Padding) {
	c.padding = p
}

func (c *Box) Render(ctx core.RenderContext) ora.Component {
	if c.with != nil {
		c.with(c)
	}

	var tmp []ora.AlignedComponent
	for _, child := range c.children {
		tmp = append(tmp, ora.AlignedComponent{
			Component: child.Component.Render(ctx),
			Alignment: child.Alignment,
		})
	}

	return ora.Box{
		Type:            ora.BoxT,
		Children:        tmp,
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Padding:         c.padding,
	}
}
