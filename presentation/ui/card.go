package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Card struct {
	id         CID
	children   *SharedList[core.Component]
	properties []core.Property
	action     *Func
}

func NewCard(with func(card *Card)) *Card {
	c := &Card{
		id:       nextPtr(),
		children: NewSharedList[core.Component]("children"),
		action:   NewFunc("action"),
	}

	c.properties = []core.Property{c.children, c.action}
	if with != nil {
		with(c)
	}
	return c
}

func (c *Card) Action() *Func {
	return c.action
}

func (c *Card) Append(children ...LiveComponent) *Card {
	c.children.Append(children...)
	return c
}

func (c *Card) ID() CID {
	return c.id
}

func (c *Card) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
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
	}
}
