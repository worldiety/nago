package uilegacy

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Breadcrumbs struct {
	id                ora.Ptr
	items             *SharedList[*BreadcrumbItem]
	selectedItemIndex Int
	icon              EmbeddedSVG
	properties        []core.Property
}

func NewBreadcrumbs(with func(breadcrumbs *Breadcrumbs)) *Breadcrumbs {
	b := &Breadcrumbs{
		id:                nextPtr(),
		items:             NewSharedList[*BreadcrumbItem]("items"),
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

func (b *Breadcrumbs) Items() *SharedList[*BreadcrumbItem] {
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
	var items []ora.BreadcrumbItem
	b.items.Iter(func(it *BreadcrumbItem) bool {
		items = append(items, it.render())
		return true
	})

	return ora.Breadcrumbs{
		Ptr:  b.id,
		Type: ora.BreadcrumbsT,
		Items: ora.Property[[]ora.BreadcrumbItem]{
			Ptr:   b.items.ID(),
			Value: items,
		},
		SelectedItemIndex: b.selectedItemIndex.render(),
		Icon:              b.icon.render(),
	}
}

type BreadcrumbItem struct {
	id         ora.Ptr
	label      String
	action     *Func
	properties []core.Property
}

func NewBreadcrumbItem(with func(breadcrumbItem *BreadcrumbItem)) *BreadcrumbItem {
	b := &BreadcrumbItem{
		id:     nextPtr(),
		label:  NewShared[string]("label"),
		action: NewFunc("action"),
	}

	b.properties = []core.Property{b.label, b.action}
	if with != nil {
		with(b)
	}
	return b
}

func (b *BreadcrumbItem) ID() ora.Ptr {
	return b.id
}

func (b *BreadcrumbItem) Label() String {
	return b.label
}

func (b *BreadcrumbItem) Action() *Func {
	return b.action
}

func (b *BreadcrumbItem) Properties(yield func(core.Property) bool) {
	for _, property := range b.properties {
		if !yield(property) {
			return
		}
	}
}

func (b *BreadcrumbItem) Render() ora.Component {
	return b.render()
}

func (b *BreadcrumbItem) render() ora.BreadcrumbItem {
	return ora.BreadcrumbItem{
		Ptr:    b.id,
		Type:   ora.BreadcrumbItemT,
		Label:  b.label.render(),
		Action: renderFunc(b.action),
	}
}
