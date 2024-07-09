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
	color      *Shared[ora.NamedColor]
	action     *Func
	disabled   Bool
	visible    Bool
	properties []core.Property
	frame      ora.Frame
}

func NewActionButton(caption string, action func()) *Button {
	return NewButton(func(btn *Button) {
		btn.Caption().Set(caption)
		btn.Action().Set(action)
	})
}

func NewButton(with func(btn *Button)) *Button {
	c := &Button{
		id:       nextPtr(),
		caption:  NewShared[string]("caption"),
		preIcon:  NewShared[SVGSrc]("preIcon"),
		postIcon: NewShared[SVGSrc]("postIcon"),
		color:    NewShared[ora.NamedColor]("color"),
		disabled: NewShared[bool]("disabled"),
		visible:  NewShared[bool]("visible"),
		action:   NewFunc("action"),
	}

	c.properties = []core.Property{c.caption, c.preIcon, c.postIcon, c.color, c.disabled, c.action, c.visible}
	c.Style().Set(ora.Primary) // the default style is undefined, so make it primary by default
	c.visible.Set(true)
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

func (c *Button) Frame() *ora.Frame {
	return &c.frame
}

func (c *Button) Caption() String {
	return c.caption
}

func (c *Button) Style() *Shared[ora.NamedColor] {
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

func (c *Button) Visible() Bool {
	return c.visible
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
		Visible:  c.visible.render(),
		Action:   renderFunc(c.action),
		Frame:    c.frame,
	}
}
