package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type ViewText struct {
	content         string
	color           ora.NamedColor
	backgroundColor ora.NamedColor
	size            int
	invisible       bool
	onClick         func()
	onHoverStart    func()
	onHoverEnd      func()
	Padding         ora.Padding
	frame           ora.Frame
	with            func(*ViewText)
}

func Text(with func(*ViewText)) *ViewText {
	c := &ViewText{
		with: with,
	}

	return c
}

func TextFrom(s string) *ViewText {
	return Text(func(text *ViewText) {
		text.Content(s)
	})
}

func (c *ViewText) Content(v string) {
	c.content = v
}

func (c *ViewText) Color(color ora.NamedColor) {
	c.color = color
}

func (c *ViewText) BackgroundColor(backgroundColor ora.NamedColor) {
	c.backgroundColor = backgroundColor
}

func (c *ViewText) Render(ctx core.RenderContext) ora.Component {
	if c.with != nil {
		c.with(c)
	}

	return ora.Text{
		Type:            ora.TextT,
		Value:           c.content,
		Color:           c.color,
		BackgroundColor: c.backgroundColor,
		Size:            "fix me",
		OnClick:         ctx.MountCallback(c.onClick),
		OnHoverStart:    ctx.MountCallback(c.onHoverStart),
		OnHoverEnd:      ctx.MountCallback(c.onHoverEnd),
		Invisible:       c.invisible,
		Padding:         c.Padding,
		Frame:           c.frame,
	}
}
