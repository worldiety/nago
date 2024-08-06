package uilegacy

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Divider struct {
	id ora.Ptr
}

func NewDivider(with func(*Divider)) *Divider {
	c := &Divider{
		id: nextPtr(),
	}

	if with != nil {
		with(c)
	}

	return c
}

func (c *Divider) ID() ora.Ptr {
	return c.id
}

func (c *Divider) Properties(yield func(core.Property) bool) {
}

func (c *Divider) Render() ora.Component {
	return ora.Divider{
		Ptr:  c.id,
		Type: ora.DividerT,
	}
}
