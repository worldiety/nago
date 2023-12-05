package ui

import "go.wdy.de/nago/container/slice"

type Button struct {
	id         CID
	caption    String
	preIcon    EmbeddedSVG
	postIcon   EmbeddedSVG
	color      *Shared[Color]
	action     *Func
	disabled   Bool
	properties slice.Slice[Property]
}

func NewButton(with func(btn *Button)) *Button {
	c := &Button{
		id:       nextPtr(),
		caption:  NewShared[string]("caption"),
		preIcon:  NewShared[SVGSrc]("preIcon"),
		postIcon: NewShared[SVGSrc]("postIcon"),
		color:    NewShared[Color]("color"),
		disabled: NewShared[bool]("disabled"),
		action:   NewFunc("action"),
	}

	c.properties = slice.Of[Property](c.caption, c.preIcon, c.postIcon, c.color, c.disabled, c.action)
	if with != nil {
		with(c)
	}
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

func (c *Button) Caption() String {
	return c.caption
}

func (c *Button) Style() *Shared[Color] {
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
