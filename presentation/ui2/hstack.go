package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type THStack struct {
	children           []core.View
	alignment          ora.Alignment
	backgroundColor    ora.Color
	frame              ora.Frame
	gap                ora.Length
	padding            ora.Padding
	font               ora.Font
	border             ora.Border
	accessibilityLabel string
	invisible          bool
}

func HStack(children ...core.View) *THStack {
	c := &THStack{
		children: children,
	}

	return c
}

func (c THStack) Padding(padding ora.Padding) core.DecoredView {
	c.padding = padding
	return c
}

func (c THStack) Gap(gap ora.Length) {
	c.gap = gap
}

func (c THStack) BackgroundColor(backgroundColor ora.Color) core.DecoredView {
	c.backgroundColor = backgroundColor
	return c
}

func (c THStack) Alignment(alignment ora.Alignment) THStack {
	c.alignment = alignment
	return c
}

func (c THStack) Frame(fr ora.Frame) core.DecoredView {
	c.frame = fr
	return c
}

func (c THStack) Font(font ora.Font) core.DecoredView {
	c.font = font
	return c
}

func (c THStack) Border(border ora.Border) core.DecoredView {
	c.border = border
	return c
}

func (c THStack) Visible(visible bool) core.DecoredView {
	c.invisible = !visible
	return c
}

func (c THStack) AccessibilityLabel(label string) core.DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c THStack) Render(ctx core.RenderContext) ora.Component {

	return ora.HStack{
		Type:               ora.HStackT,
		Children:           renderComponents(ctx, c.children),
		Gap:                c.gap,
		Frame:              c.frame,
		Alignment:          c.alignment,
		BackgroundColor:    c.backgroundColor,
		Padding:            c.padding,
		Border:             c.border,
		AccessibilityLabel: c.accessibilityLabel,
		Invisible:          c.invisible,
		Font:               c.font,
	}
}
