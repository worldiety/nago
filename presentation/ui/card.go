package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Card struct {
	id         ora.Ptr
	children   *SharedList[core.View]
	properties []core.Property
	visible    Bool
	action     *Func
}

func NewCard(with func(card *Card)) *Card {
	c := &Card{
		id:       nextPtr(),
		children: NewSharedList[core.View]("children"),
		action:   NewFunc("action"),
		visible:  NewShared[bool]("visible"),
	}

	c.properties = []core.Property{c.children, c.action, c.visible}
	c.visible.Set(true)
	if with != nil {
		with(c)
	}
	return c
}

func (c *Card) Action() *Func {
	return c.action
}

func (c *Card) Append(children ...core.View) *Card {
	c.children.Append(children...)
	return c
}

func (c *Card) Children() *SharedList[core.View] {
	return c.children
}

func (c *Card) ID() ora.Ptr {
	return c.id
}

func (c *Card) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Card) Visible() Bool {
	return c.visible
}

func (c *Card) Render() ora.Component {
	return c.render()
}

func (c *Card) render() ora.Card {
	return ora.Card{
		Ptr:      c.id,
		Type:     ora.CardT,
		Children: renderSharedListComponents(c.children),
		Action:   renderFunc(c.action),
		Visible:  c.visible.render(),
	}
}
