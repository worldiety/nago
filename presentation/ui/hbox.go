package ui

import (
	"go.wdy.de/nago/container/slice"
)

type HBox struct {
	id         CID
	children   *SharedList[LiveComponent]
	alignment  String
	properties slice.Slice[Property]
}

func NewHBox(with func(hbox *HBox)) *HBox {
	c := &HBox{
		id:        nextPtr(),
		children:  NewSharedList[LiveComponent]("children"),
		alignment: NewShared[string]("alignment"),
	}
	c.alignment.Set("grid")
	c.alignment.SetDirty(false)

	c.properties = slice.Of[Property](c.children, c.alignment)
	if with != nil {
		with(c)
	}

	return c
}

func (c *HBox) Append(children ...LiveComponent) *HBox {
	c.children.Append(children...)
	return c
}

func (c *HBox) Children() *SharedList[LiveComponent] {
	return c.children
}

// Alignment is a layout hint. Supported values are "grid" (default) or "flex-left", "flex-right" or "flex-center".
func (c *HBox) Alignment() String {
	return c.alignment
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
