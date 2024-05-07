package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Breadcrumbs struct {
	id                ora.Ptr
	items             *SharedList[string]
	selectedItemIndex Int
	icon              EmbeddedSVG
	properties        []core.Property
}

func NewBreadcrumbs(with func(dropdown *Breadcrumbs)) *Breadcrumbs {
	b := &Breadcrumbs{
		id:                nextPtr(),
		items:             NewSharedList[string]("items"),
		selectedItemIndex: NewShared[int64]("selectedItemIndex"),
		icon:              NewShared[SVGSrc]("icon"),
	}

	b.properties = []core.Property{b.items, b.selectedItemIndex, b.icon}
	if with != nil {
		with(b)
	}
	return b
}

func (b *Breadcrumbs) ID() ora.Ptr {
	return b.id
}

func (b *Breadcrumbs) Items() *SharedList[string] {
	return b.items
}

func (b *Breadcrumbs) SelectedItemIndex() Int {
	return b.selectedItemIndex
}

func (b *Breadcrumbs) Icon() EmbeddedSVG {
	return b.icon
}

func (b *Breadcrumbs) Properties(yield func(core.Property) bool) {
	for _, property := range b.properties {
		if !yield(property) {
			return
		}
	}
}

func (b *Breadcrumbs) Render() ora.Component {
	return b.render()
}

func (b *Breadcrumbs) render() ora.Breadcrumbs {
	return ora.Breadcrumbs{
		Ptr:               b.id,
		Type:              ora.BreadcrumbsT,
		Items:             b.items.render(),
		SelectedItemIndex: b.selectedItemIndex.render(),
		Icon:              b.icon.render(),
	}
}
