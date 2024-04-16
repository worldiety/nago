package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

// A Chip is like a badge but removable.
type Chip struct {
	id         CID
	caption    String
	action     *Func
	onClose    *Func
	color      *Shared[Color]
	properties []core.Property
}

func NewChip(with func(chip *Chip)) *Chip {
	c := &Chip{
		id:      nextPtr(),
		caption: NewShared[string]("caption"),
		action:  NewFunc("action"),
		onClose: NewFunc("onClose"),
		color:   NewShared[Color]("color"),
	}

	c.properties = []core.Property{c.caption, c.action, c.onClose, c.color}

	if with != nil {
		with(c)
	}

	return c
}

func (c *Chip) ID() CID {
	return c.id
}

func (c *Chip) Caption() String {
	return c.caption
}

func (c *Chip) Action() *Func {
	return c.action
}

func (c *Chip) OnClose() *Func {
	return c.onClose
}

// TODO TBD: red, green, yellow
func (c *Chip) Color() *Shared[Color] {
	return c.color
}

func (c *Chip) Properties(yield func(property core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Chip) Render() ora.Component {
	return c.render()
}

func (c *Chip) render() ora.Chip {
	return ora.Chip{
		Ptr:     c.id,
		Type:    ora.ChipT,
		Caption: c.caption.render(),
		Action:  renderFunc(c.action),
		OnClose: renderFunc(c.onClose),
		Color: ora.Property[string]{
			Ptr:   c.color.id,
			Value: string(c.color.v),
		},
	}
}
