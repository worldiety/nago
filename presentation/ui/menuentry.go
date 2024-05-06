package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type MenuEntry struct {
	id         ora.Ptr
	icon       EmbeddedSVG
	iconActive EmbeddedSVG
	title      String
	action     *Func
	menu       *SharedList[*MenuEntry]
	expanded   Bool
	onFocus    *Func
	properties []core.Property
}

func NewMenuEntry(with func(menuEntry *MenuEntry)) *MenuEntry {
	m := &MenuEntry{
		id:         nextPtr(),
		icon:       NewShared[SVGSrc]("icon"),
		iconActive: NewShared[SVGSrc]("iconActive"),
		title:      NewShared[string]("title"),
		action:     NewFunc("action"),
		menu:       NewSharedList[*MenuEntry]("menu"),
		expanded:   NewShared[bool]("expanded"),
		onFocus:    NewFunc("onFocus"),
	}

	m.properties = []core.Property{m.icon, m.iconActive, m.title, m.action, m.menu, m.expanded, m.onFocus}
	if with != nil {
		with(m)
	}
	return m
}

func (m *MenuEntry) ID() ora.Ptr {
	return m.id
}

func (c *MenuEntry) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
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

func (m *MenuEntry) Icon() EmbeddedSVG {
	return m.icon
}

func (m *MenuEntry) IconActive() EmbeddedSVG {
	return m.iconActive
}

func (m *MenuEntry) Menu() *SharedList[*MenuEntry] {
	return m.menu
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
		Ptr:        m.id,
		Type:       ora.MenuEntryT,
		Title:      m.title.render(),
		Action:     renderFunc(m.action),
		Icon:       m.icon.render(),
		IconActive: m.iconActive.render(),
		Menu:       renderSharedListMenuEntries(m.menu),
		Expanded:   m.expanded.render(),
		OnFocus:    renderFunc(m.onFocus),
	}
}
