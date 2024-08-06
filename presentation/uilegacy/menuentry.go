package uilegacy

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type MenuEntry struct {
	id                 ora.Ptr
	icon               EmbeddedSVG
	iconActive         EmbeddedSVG
	title              String
	action             *Func
	componentFactoryId String
	menu               *SharedList[*MenuEntry]
	badge              String
	expanded           Bool
	onFocus            *Func
	properties         []core.Property
}

func NewMenuEntry(with func(menuEntry *MenuEntry)) *MenuEntry {
	m := &MenuEntry{
		id:                 nextPtr(),
		icon:               NewShared[SVGSrc]("icon"),
		iconActive:         NewShared[SVGSrc]("iconActive"),
		title:              NewShared[string]("title"),
		action:             NewFunc("action"),
		componentFactoryId: NewShared[string]("componentFactoryId"),
		menu:               NewSharedList[*MenuEntry]("menu"),
		badge:              NewShared[string]("badge"),
		expanded:           NewShared[bool]("expanded"),
		onFocus:            NewFunc("onFocus"),
	}

	m.properties = []core.Property{m.icon, m.iconActive, m.title, m.action, m.componentFactoryId, m.menu, m.badge, m.expanded, m.onFocus}
	if with != nil {
		with(m)
	}
	return m
}

func (m *MenuEntry) ID() ora.Ptr {
	return m.id
}

func (m *MenuEntry) Properties(yield func(core.Property) bool) {
	for _, property := range m.properties {
		if !yield(property) {
			return
		}
	}
}

func (m *MenuEntry) Title() String {
	return m.title
}

func (m *MenuEntry) Action() *Func {
	return m.action
}

func (m *MenuEntry) ComponentFactoryId() String {
	return m.componentFactoryId
}

func (m *MenuEntry) Icon() EmbeddedSVG {
	return m.icon
}

func (m *MenuEntry) IconActive() EmbeddedSVG {
	return m.iconActive
}

func (m *MenuEntry) Menu() *SharedList[*MenuEntry] {
	return m.menu
}

func (m *MenuEntry) Badge() String {
	return m.badge
}

func (m *MenuEntry) Expanded() Bool {
	return m.expanded
}

func (m *MenuEntry) OnFocus() *Func {
	return m.onFocus
}

func (m *MenuEntry) Render() ora.Component {
	return m.renderMenuEntry()
}

func (m *MenuEntry) renderMenuEntry() ora.MenuEntry {
	return ora.MenuEntry{
		Ptr:                m.id,
		Type:               ora.MenuEntryT,
		Title:              m.title.render(),
		Action:             renderFunc(m.action),
		ComponentFactoryId: m.componentFactoryId.render(),
		Icon:               m.icon.render(),
		IconActive:         m.iconActive.render(),
		Menu:               renderSharedListMenuEntries(m.menu),
		Badge:              m.badge.render(),
		Expanded:           m.expanded.render(),
		OnFocus:            renderFunc(m.onFocus),
	}
}

func (m *MenuEntry) Link(componentFactoryId ora.ComponentFactoryId, window core.Window, query core.Values) {
	m.componentFactoryId.Set(string(componentFactoryId))
	m.action.Set(func() {
		window.Navigation().ForwardTo(componentFactoryId, query)
	})
}
