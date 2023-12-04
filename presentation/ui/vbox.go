package ui

import (
	"go.wdy.de/nago/container/slice"
)

type VBox struct {
	id         CID
	children   *SharedList[LiveComponent]
	properties slice.Slice[Property]
}

func NewVBox(with func(vbox *VBox)) *VBox {
	c := &VBox{
		id:       nextPtr(),
		children: NewSharedList[LiveComponent]("children"),
	}

	c.properties = slice.Of[Property](c.children)
	if with != nil {
		with(c)
	}
	return c
}

func (c *VBox) Append(children ...LiveComponent) *VBox {
	c.children.Append(children...)
	return c
}

func (c *VBox) ID() CID {
	return c.id
}

func (c *VBox) Type() string {
	return "VBox"
}

func (c *VBox) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *VBox) Children() slice.Slice[LiveComponent] {
	return slice.Of(c.children.values...)
}

func (c *VBox) Functions() slice.Slice[*Func] {
	return slice.Of[*Func]()
}
