package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type HStack struct {
	children        []core.View
	alignment       ora.Alignment
	backgroundColor ora.NamedColor
	frame           ora.Frame
	gap             ora.Length
	padding         ora.Padding
	with            func(stack *HStack)
}

func NewHStack(with func(hstack *HStack)) *HStack {
	c := &HStack{
		with: with,
	}

	c.alignment = "" // if nothing is defined, ora.Center must be applied by renderer

	return c
}

func (c *HStack) Padding(padding ora.Padding) {
	c.padding = padding
}

func (c *HStack) Gap(gap ora.Length) {
	c.gap = gap
}

func (c *HStack) BackgroundColor(backgroundColor ora.NamedColor) {
	c.backgroundColor = backgroundColor
}

func (c *HStack) Alignment(alignment ora.Alignment) {
	c.alignment = alignment
}

func (c *HStack) Append(children ...core.View) {
	// this signature does not return builder pattern anymore, because it makes polymorphic interface usage impossible
	c.children = append(c.children, children...)
}

func (c *HStack) Frame(fr ora.Frame) {
	c.frame = fr
}

func (c *HStack) Render(ctx core.RenderContext) ora.Component {
	if c.with != nil {
		c.with(c)
	}

	return ora.HStack{
		Type:            ora.HStackT,
		Children:        renderComponents(ctx, c.children),
		Frame:           c.frame,
		Alignment:       c.alignment,
		BackgroundColor: c.backgroundColor,
		Gap:             c.gap,
		Padding:         c.padding,
	}
}
