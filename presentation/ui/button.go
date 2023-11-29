package ui

import "go.wdy.de/nago/container/slice"

type Button struct {
	id         CID
	caption    String
	preIcon    EmbeddedSVG
	postIcon   EmbeddedSVG
	color      Color
	action     *Func
	disabled   Bool
	properties slice.Slice[Property]
	functions  slice.Slice[*Func]
}

func NewButton() *Button {
	c := &Button{
		id:       nextPtr(),
		caption:  NewShared[string]("caption"),
		preIcon:  NewShared[SVGSrc]("preIcon"),
		postIcon: NewShared[SVGSrc]("postIcon"),
		color:    NewShared[IntentColor]("color"),
		disabled: NewShared[bool]("disabled"),
		action:   NewFunc("action"),
	}

	c.properties = slice.Of[Property](c.caption, c.preIcon, c.postIcon, c.color, c.disabled, c.action)
	c.functions = slice.Of[*Func](c.action)
	return c
}

func (c *Button) With(f func(btn *Button)) *Button {
	f(c)
	return c
}

func (c *Button) ID() CID {
	return c.id
}

func (c *Button) Type() string {
	return "Button"
}

func (c *Button) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *Button) Children() slice.Slice[LiveComponent] {
	return slice.Of[LiveComponent]()
}

func (c *Button) Functions() slice.Slice[*Func] {
	return c.functions
}

func (c *Button) Caption() String {
	return c.caption
}

func (c *Button) Style() Color {
	return c.color
}

func (c *Button) PreIcon() EmbeddedSVG {
	return c.preIcon
}

func (c *Button) PostIcon() EmbeddedSVG {
	return c.postIcon
}

func (c *Button) Action() *Func {
	return c.action
}

func (c *Button) Disabled() Bool {
	return c.disabled
}
