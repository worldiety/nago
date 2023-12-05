package ui

import (
	"go.wdy.de/nago/container/slice"
)

type Card struct {
	id         CID
	children   *SharedList[LiveComponent]
	properties slice.Slice[Property]
	action     *Func
}

func NewCard(with func(card *Card)) *Card {
	c := &Card{
		id:       nextPtr(),
		children: NewSharedList[LiveComponent]("children"),
		action:   NewFunc("action"),
	}

	c.properties = slice.Of[Property](c.children, c.action)
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

func (c *Card) Type() string {
	return "Card"
}

func (c *Card) Properties() slice.Slice[Property] {
	return c.properties
}
