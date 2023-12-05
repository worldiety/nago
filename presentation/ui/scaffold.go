package ui

import "go.wdy.de/nago/container/slice"

type Scaffold struct {
	id          CID
	title       String
	breadcrumbs *SharedList[*Button]
	menu        *SharedList[*Button]
	body        *Shared[LiveComponent]
	topbarLeft  *Shared[LiveComponent]
	topbarMid   *Shared[LiveComponent]
	topbarRight *Shared[LiveComponent]
	properties  slice.Slice[Property]
}

func NewScaffold(with func(scaffold *Scaffold)) *Scaffold {
	c := &Scaffold{
		id:          nextPtr(),
		title:       NewShared[string]("title"),
		breadcrumbs: NewSharedList[*Button]("breadcrumbs"),
		topbarLeft:  NewShared[LiveComponent]("topbarLeft"),
		topbarMid:   NewShared[LiveComponent]("topbarMid"),
		topbarRight: NewShared[LiveComponent]("topbarRight"),
		menu:        NewSharedList[*Button]("menu"),
		body:        NewShared[LiveComponent]("body"),
	}

	c.properties = slice.Of[Property](c.title, c.breadcrumbs, c.menu, c.body, c.topbarLeft, c.topbarMid, c.topbarRight)

	if with != nil {
		with(c)
	}

	return c
}

func (c *Scaffold) Body() *Shared[LiveComponent] {
	return c.body
}

func (c *Scaffold) Menu() *SharedList[*Button] {
	return c.menu
}

func (c *Scaffold) Breadcrumbs() *SharedList[*Button] {
	return c.breadcrumbs
}

func (c *Scaffold) TopBar() ScaffoldTopBar {
	return ScaffoldTopBar{
		Left:  c.topbarLeft,
		Mid:   c.topbarMid,
		Right: c.topbarRight,
	}
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

type ScaffoldTopBar struct {
	Left  *Shared[LiveComponent]
	Mid   *Shared[LiveComponent]
	Right *Shared[LiveComponent]
}
