package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type Scaffold struct {
	id          CID
	title       String
	breadcrumbs *SharedList[*Button]
	menu        *SharedList[*Button]
	body        *Shared[core.Component]
	topbarLeft  *Shared[core.Component]
	topbarMid   *Shared[core.Component]
	topbarRight *Shared[core.Component]
	properties  []core.Property
}

func NewScaffold(with func(scaffold *Scaffold)) *Scaffold {
	c := &Scaffold{
		id:          nextPtr(),
		title:       NewShared[string]("title"),
		breadcrumbs: NewSharedList[*Button]("breadcrumbs"),
		topbarLeft:  NewShared[core.Component]("topbarLeft"),
		topbarMid:   NewShared[core.Component]("topbarMid"),
		topbarRight: NewShared[core.Component]("topbarRight"),
		menu:        NewSharedList[*Button]("menu"),
		body:        NewShared[core.Component]("body"),
	}

	c.properties = []core.Property{c.title, c.breadcrumbs, c.menu, c.body, c.topbarLeft, c.topbarMid, c.topbarRight}

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

func (c *Scaffold) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Scaffold) Render() protocol.Component {
	return protocol.Scaffold{
		Ptr:         c.id,
		Type:        protocol.ScaffoldT,
		Title:       c.title.render(),
		Body:        renderComponentProp(c.body, c.body),
		Breadcrumbs: renderSharedListButtons(c.breadcrumbs),
		Menu:        renderSharedListButtons(c.menu),
		TopbarLeft:  renderSharedComponent(c.topbarLeft),
		TopbarMid:   renderSharedComponent(c.topbarMid),
		TopbarRight: renderSharedComponent(c.topbarRight),
	}
}

type ScaffoldTopBar struct {
	Left  *Shared[core.Component]
	Mid   *Shared[core.Component]
	Right *Shared[core.Component]
}
