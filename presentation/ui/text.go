package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Text struct {
	id           CID
	value        String
	color        *Shared[Color]
	colorDark    *Shared[Color]
	size         *Shared[Size]
	properties   []core.Property
	onClick      *Func
	onHoverStart *Func
	onHoverEnd   *Func
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
	c.color = NewShared[Color]("color")
	c.colorDark = NewShared[Color]("colorDark")
	c.properties = []core.Property{c.value, c.color, c.colorDark, c.size, c.onClick, c.onHoverStart, c.onHoverEnd}
	if with != nil {
		with(c)
	}

	return c
}

func MakeText(s string) *Text {
	return NewText(func(text *Text) {
		text.Value().Set(s)
	})
}

func (c *Text) Value() String {
	return c.value
}

func (c *Text) Color() *Shared[Color] {
	return c.color
}

func (c *Text) ColorDark() *Shared[Color] {
	return c.colorDark
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

func (c *Text) ID() CID {
	return c.id
}

func (c *Text) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Text) Render() ora.Component {
	return ora.Text{
		Ptr:   c.id,
		Type:  ora.TextT,
		Value: c.value.render(),
		Color: ora.Property[string]{
			Ptr:   c.color.id,
			Value: string(c.color.v),
		},
		ColorDark: ora.Property[string]{
			Ptr:   c.colorDark.id,
			Value: string(c.colorDark.v),
		},
		Size: ora.Property[string]{
			Ptr:   c.size.id,
			Value: string(c.size.v),
		},
		OnClick:      renderFunc(c.onClick),
		OnHoverStart: renderFunc(c.onHoverStart),
		OnHoverEnd:   renderFunc(c.onHoverEnd),
	}
}
