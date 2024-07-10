package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Text struct {
	id              ora.Ptr
	value           String
	color           ora.NamedColor
	backgroundColor ora.NamedColor
	size            *Shared[Size]
	visible         Bool
	properties      []core.Property
	onClick         *Func
	onHoverStart    *Func
	onHoverEnd      *Func
	Padding         ora.Padding
	Frame           ora.Frame
}

func NewText(with func(*Text)) *Text {
	c := &Text{
		id: nextPtr(),
	}

	c.value = NewShared[string]("value")
	c.onClick = NewFunc("onClick")
	c.onHoverStart = NewFunc("onHoverStart")
	c.onHoverEnd = NewFunc("onHoverEnd")
	c.size = NewShared[Size]("size")
	c.visible = NewShared[bool]("visible")
	c.properties = []core.Property{c.value, c.size, c.onClick, c.onHoverStart, c.onHoverEnd, c.visible}
	c.visible.Set(true)
	if with != nil {
		with(c)
	}

	return c
}

// deprecated: use NewStr instead which is a lot cheaper.
func MakeText(s string) *Text {
	return NewText(func(text *Text) {
		text.Value().Set(s)
	})
}

func (c *Text) Value() String {
	return c.value
}

func (c *Text) Color() ora.NamedColor {
	return c.color
}

func (c *Text) SetColor(color ora.NamedColor) {
	c.color = color
}

func (c *Text) BackgroundColor() ora.NamedColor {
	return c.backgroundColor
}

func (c *Text) SetBackgroundColor(backgroundColor ora.NamedColor) {
	c.backgroundColor = backgroundColor
}

func (c *Text) Size() *Shared[Size] {
	return c.size
}

func (c *Text) OnClick() *Func {
	return c.onClick
}

func (c *Text) OnHoverStart() *Func {
	return c.onHoverStart
}

func (c *Text) OnHoverEnd() *Func {
	return c.onHoverEnd
}

func (c *Text) ID() ora.Ptr {
	return c.id
}

func (c *Text) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Text) Visible() Bool {
	return c.visible
}

func (c *Text) Render() ora.Component {
	return ora.Text{
		Ptr:             c.id,
		Type:            ora.TextT,
		Value:           c.value.render(),
		Color:           c.color,
		BackgroundColor: c.backgroundColor,
		Size: ora.Property[string]{
			Ptr:   c.size.id,
			Value: string(c.size.v),
		},
		OnClick:      renderFunc(c.onClick),
		OnHoverStart: renderFunc(c.onHoverStart),
		OnHoverEnd:   renderFunc(c.onHoverEnd),
		Visible:      c.visible.render(),
		Padding:      c.Padding,
		Frame:        c.Frame,
	}
}
