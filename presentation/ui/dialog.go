package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type ModalOwner interface {
	Modals() *SharedList[core.Component]
}

type Dialog struct {
	id      ora.Ptr
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

func (c *Dialog) ID() ora.Ptr {
	return c.id
}

func (c *Dialog) Type() ora.ComponentType {
	return ora.DialogT
}

func (c *Dialog) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Dialog) Render() ora.Component {
	return c.render()
}

func (c *Dialog) render() ora.Dialog {
	return ora.Dialog{
		Ptr:     c.id,
		Type:    ora.DialogT,
		Title:   c.title.render(),
		Body:    renderSharedComponent(c.body),
		Icon:    c.icon.render(),
		Actions: renderSharedListButtons(c.actions),
	}
}
