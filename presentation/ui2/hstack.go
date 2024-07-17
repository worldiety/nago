package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type THStack struct {
	children        []core.View
	alignment       ora.Alignment
	backgroundColor ora.Color
	frame           ora.Frame
	gap             ora.Length
	padding         ora.Padding
	font            ora.Font
}

func HStack(children ...core.View) *THStack {
	c := &THStack{
		children: children,
	}

	return c
}

func (c THStack) Padding(padding ora.Padding) {
	c.padding = padding
}

func (c THStack) Gap(gap ora.Length) {
	c.gap = gap
}

func (c THStack) BackgroundColor(backgroundColor ora.Color) {
	c.backgroundColor = backgroundColor
}

func (c THStack) Alignment(alignment ora.Alignment) {
	c.alignment = alignment
}

func (c THStack) Frame(fr ora.Frame) THStack {
	c.frame = fr
	return c
}

func (c THStack) Font(font ora.Font) THStack {
	c.font = font
	return c
}

func (c THStack) Render(ctx core.RenderContext) ora.Component {

	return ora.HStack{
		Type:            ora.HStackT,
		Children:        renderComponents(ctx, c.children),
		Frame:           c.frame,
		Alignment:       c.alignment,
		BackgroundColor: c.backgroundColor,
		Gap:             c.gap,
		Padding:         c.padding,
		Font:            c.font,
	}
}
