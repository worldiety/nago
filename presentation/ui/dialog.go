package ui

import "go.wdy.de/nago/container/slice"

type Dialog struct {
	id      CID
	title   String
	body    *Shared[LiveComponent]
	icon    *Shared[SVGSrc]
	actions *SharedList[*Button]

	properties slice.Slice[Property]
}

func NewDialog(with func(dlg *Dialog)) *Dialog {
	c := &Dialog{
		id:      nextPtr(),
		title:   NewShared[string]("title"),
		icon:    NewShared[SVGSrc]("icon"),
		body:    NewShared[LiveComponent]("body"),
		actions: NewSharedList[*Button]("actions"),
	}

	c.properties = slice.Of[Property](c.title, c.icon, c.body, c.actions)

	if with != nil {
		with(c)
	}
	return c
}

func (c *Dialog) Title() String {
	return c.title
}

func (c *Dialog) Body() *Shared[LiveComponent] {
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

func (c *Dialog) Type() string {
	return "Dialog"
}

func (c *Dialog) Properties() slice.Slice[Property] {
	return c.properties
}
