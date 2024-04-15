package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type Dialog struct {
	id      CID
	title   String
	body    *Shared[core.Component]
	icon    *Shared[SVGSrc]
	actions *SharedList[*Button]

	properties []core.Property
}

func NewDialog(with func(dlg *Dialog)) *Dialog {
	c := &Dialog{
		id:      nextPtr(),
		title:   NewShared[string]("title"),
		icon:    NewShared[SVGSrc]("icon"),
		body:    NewShared[core.Component]("body"),
		actions: NewSharedList[*Button]("actions"),
	}

	c.properties = []core.Property{c.title, c.icon, c.body, c.actions}

	if with != nil {
		with(c)
	}
	return c
}

func (c *Dialog) Title() String {
	return c.title
}

func (c *Dialog) Body() *Shared[core.Component] {
	return c.body
}

func (c *Dialog) Icon() *Shared[SVGSrc] {
	return c.icon
}

func (c *Dialog) Actions() *SharedList[*Button] {
	return c.actions
}

func (c *Dialog) ID() CID {
	return c.id
}

func (c *Dialog) Type() protocol.ComponentType {
	return protocol.DialogT
}

func (c *Dialog) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Dialog) Render() protocol.Component {
	panic("implement me")
}
