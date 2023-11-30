package ui

import (
	"go.wdy.de/nago/container/slice"
)

type Divider struct {
	id         CID
	children   *SharedList[LiveComponent]
	properties slice.Slice[Property]
}

func NewDivider(with func(*Divider)) *Divider {
	c := &Divider{
		id:       nextPtr(),
		children: NewSharedList[LiveComponent]("children"),
	}

	c.properties = slice.Of[Property](c.children)

	if with != nil {
		with(c)
	}

	return c
}

func (c *Divider) ID() CID {
	return c.id
}

func (c *Divider) Type() string {
	return "Divider"
}

func (c *Divider) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *Divider) Children() slice.Slice[LiveComponent] {
	return slice.Of(c.children.values...)
}

func (c *Divider) Functions() slice.Slice[*Func] {
	return slice.Of[*Func]()
}
