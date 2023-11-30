package ui

import (
	"go.wdy.de/nago/container/slice"
)

type HBox struct {
	id         CID
	children   *SharedList[LiveComponent]
	properties slice.Slice[Property]
}

func NewHBox(with func(box *HBox)) *HBox {
	c := &HBox{
		id:       nextPtr(),
		children: NewSharedList[LiveComponent]("children"),
	}

	c.properties = slice.Of[Property](c.children)
	if with != nil {
		with(c)
	}

	return c
}

func (c *HBox) Append(children ...LiveComponent) *HBox {
	c.children.Append(children...)
	return c
}

func (c *HBox) ID() CID {
	return c.id
}

func (c *HBox) Type() string {
	return "HBox"
}

func (c *HBox) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *HBox) Children() slice.Slice[LiveComponent] {
	return slice.Of(c.children.values...)
}

func (c *HBox) Functions() slice.Slice[*Func] {
	return slice.Of[*Func]()
}
