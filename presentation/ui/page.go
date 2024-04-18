package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Page struct {
	id         ora.Ptr
	body       *Shared[core.Component]
	modals     *SharedList[core.Component]
	properties []core.Property
}

func NewPage(with func(page *Page)) *Page {
	p := &Page{id: nextPtr()}
	p.body = NewShared[core.Component]("body")
	p.modals = NewSharedList[core.Component]("modals")
	p.properties = []core.Property{p.body, p.modals}
	if with != nil {
		with(p)
	}
	return p
}

func (p *Page) ID() ora.Ptr {
	return p.id
}

func (p *Page) Properties(yield func(core.Property) bool) {
	for _, property := range p.properties {
		if !yield(property) {
			return
		}
	}
}

func (p *Page) Render() ora.Component {
	return ora.Page{
		Ptr:    p.id,
		Type:   ora.PageT,
		Body:   renderComponentProp(p.body, p.body),
		Modals: renderComponentsProp(p.modals, p.modals),
	}
}

func (p *Page) Body() *Shared[core.Component] {
	return p.body
}

func (p *Page) Modals() *SharedList[core.Component] {
	return p.modals
}
