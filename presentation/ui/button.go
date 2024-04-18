package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Button struct {
	id         ora.Ptr
	caption    String
	preIcon    EmbeddedSVG
	postIcon   EmbeddedSVG
	color      *Shared[Color]
	action     *Func
	disabled   Bool
	properties []core.Property
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

	c.properties = []core.Property{c.caption, c.preIcon, c.postIcon, c.color, c.disabled, c.action}
	if with != nil {
		with(c)
	}
	return c
}

func (c *Button) ID() ora.Ptr {
	return c.id
}

func (c *Button) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
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

func (c *Button) Render() ora.Component {
	return c.renderButton()
}

func (c *Button) renderButton() ora.Button {
	return ora.Button{
		Ptr:      c.id,
		Type:     ora.ButtonT,
		Caption:  c.caption.render(),
		PreIcon:  c.preIcon.render(),
		PostIcon: c.postIcon.render(),
		Color:    c.color.render(),
		Disabled: c.disabled.render(),
		Action:   renderFunc(c.action),
	}
}
