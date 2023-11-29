package ui

import "go.wdy.de/nago/container/slice"

type Scaffold struct {
	id          CID
	title       String
	breadcrumbs *SharedList[*Button]
	menu        *SharedList[*Button]
	body        *Shared[LiveComponent]
	properties  slice.Slice[Property]
}

func NewScaffold() *Scaffold {
	c := &Scaffold{
		id:          nextPtr(),
		title:       NewShared[string]("title"),
		breadcrumbs: NewSharedList[*Button]("breadcrumbs"),
		menu:        NewSharedList[*Button]("menu"),
		body:        NewShared[LiveComponent]("body"),
	}

	c.properties = slice.Of[Property](c.title, c.breadcrumbs, c.menu, c.body)

	return c
}

func (c *Scaffold) ID() CID {
	return c.id
}

func (c *Scaffold) Type() string {
	return "Scaffold"
}

func (c *Scaffold) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *Scaffold) Children() slice.Slice[LiveComponent] {
	tmp := make([]LiveComponent, 0, c.breadcrumbs.Len()+c.menu.Len()+1)
	c.breadcrumbs.Each(func(b *Button) {
		tmp = append(tmp, b)
	})
	c.menu.Each(func(b *Button) {
		tmp = append(tmp, b)
	})
	if c.body != nil {
		tmp = append(tmp, c.body.Get())
	}

	return slice.Of[LiveComponent](tmp...)
}

func (c *Scaffold) Functions() slice.Slice[*Func] {
	return slice.Of[*Func]()
}
