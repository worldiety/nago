package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TText struct {
	content            string
	color              ora.Color
	backgroundColor    ora.Color
	size               ora.Length
	invisible          bool
	onClick            func()
	onHoverStart       func()
	onHoverEnd         func()
	padding            ora.Padding
	frame              ora.Frame
	border             ora.Border
	accessibilityLabel string
}

func Text(content string) TText {
	return TText{content: content}
}

func (c TText) Padding(padding ora.Padding) core.DecoredView {
	c.padding = padding
	return c
}

func (c TText) Frame(frame ora.Frame) core.DecoredView {
	c.frame = frame
	return c
}

func (c TText) Border(border ora.Border) core.DecoredView {
	c.border = border
	return c
}

func (c TText) Visible(visible bool) core.DecoredView {
	c.invisible = !visible
	return c
}

func (c TText) AccessibilityLabel(label string) core.DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TText) Size(size ora.Length) TText {
	c.size = size
	return c
}

func (c TText) Color(color ora.Color) TText {
	c.color = color
	return c
}

func (c TText) BackgroundColor(backgroundColor ora.Color) core.DecoredView {
	c.backgroundColor = backgroundColor
	return c
}

func (c TText) Render(ctx core.RenderContext) ora.Component {

	return ora.Text{
		Type:               ora.TextT,
		Value:              c.content,
		Color:              c.color,
		BackgroundColor:    c.backgroundColor,
		Size:               c.size,
		OnClick:            ctx.MountCallback(c.onClick),
		OnHoverStart:       ctx.MountCallback(c.onHoverStart),
		OnHoverEnd:         ctx.MountCallback(c.onHoverEnd),
		Invisible:          c.invisible,
		Padding:            c.padding,
		Frame:              c.frame,
		AccessibilityLabel: c.accessibilityLabel,
	}
}
