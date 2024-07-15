package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type ViewVStack struct {
	children        []core.View
	alignment       ora.Alignment
	backgroundColor ora.NamedColor
	frame           ora.Frame
	gap             ora.Length
	padding         ora.Padding
	with            func(stack *ViewVStack)
}

func VStack(with func(hstack *ViewVStack)) *ViewVStack {
	c := &ViewVStack{}

	c.alignment = "" // if nothing is defined, ora.Center must be applied by renderer
	c.with = with

	return c
}

func (c *ViewVStack) Gap(gap ora.Length) {
	c.gap = gap
}

func (c *ViewVStack) BackgroundColor(backgroundColor ora.NamedColor) {
	c.backgroundColor = backgroundColor
}

func (c *ViewVStack) Alignment(alignment ora.Alignment) {
	c.alignment = alignment
}

func (c *ViewVStack) Frame(f ora.Frame) {
	c.frame = f
}

func (c *ViewVStack) Of(v ...core.View) {
	c.children = append(c.children, v...)
}

func (c *ViewVStack) Render(ctx core.RenderContext) ora.Component {
	reset(&c.children)
	if c.with != nil {
		c.with(c)
	}

	return ora.VStack{
		Type:            ora.VStackT,
		Children:        renderComponents(ctx, c.children),
		Frame:           c.frame,
		Alignment:       c.alignment,
		BackgroundColor: c.backgroundColor,
		Gap:             c.gap,
		Padding:         c.padding,
	}
}
