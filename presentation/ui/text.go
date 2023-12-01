package ui

import "go.wdy.de/nago/container/slice"

type Text struct {
	id           CID
	value        String
	color        *Shared[Color]
	colorDark    *Shared[Color]
	size         *Shared[Size]
	properties   slice.Slice[Property]
	functions    slice.Slice[*Func]
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
	c.properties = slice.Of[Property](c.value, c.color, c.colorDark, c.size, c.onClick, c.onHoverStart, c.onHoverEnd)
	c.functions = slice.Of[*Func](c.onClick, c.onHoverStart, c.onHoverEnd)
	if with != nil {
		with(c)
	}

	return c
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

func (c *Text) Type() string {
	return "Text"
}

func (c *Text) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *Text) Children() slice.Slice[LiveComponent] {
	return slice.Of[LiveComponent]()
}

func (c *Text) Functions() slice.Slice[*Func] {
	return c.functions
}
