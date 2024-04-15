package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type Divider struct {
	id CID
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

func (c *Divider) ID() CID {
	return c.id
}

func (c *Divider) Properties(yield func(core.Property) bool) {
}

func (c *Divider) Render() protocol.Component {
	panic("not implemented")
}
