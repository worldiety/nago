package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type MenuEntry struct {
	id         ora.Ptr
	icon       EmbeddedSVG
	title      String
	menu       *SharedList[*MenuEntry]
	properties []core.Property
}

func NewMenuEntry(with func(menuEntry *MenuEntry)) *MenuEntry {
	m := &MenuEntry{
		id:    nextPtr(),
		icon:  NewShared[SVGSrc]("icon"),
		title: NewShared[string]("title"),
		menu:  NewSharedList[*MenuEntry]("menu"),
	}

	m.properties = []core.Property{m.icon, m.title, m.menu}
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

func (m *MenuEntry) Icon() EmbeddedSVG {
	return m.icon
}

func (m *MenuEntry) Menu() *SharedList[*MenuEntry] {
	return m.menu
}

func (m *MenuEntry) Render() ora.Component {
	return m.renderMenuEntry()
}

func (m *MenuEntry) renderMenuEntry() ora.MenuEntry {
	return ora.MenuEntry{
		Ptr:   m.id,
		Type:  ora.MenuEntryT,
		Title: m.title.render(),
		Icon:  m.icon.render(),
		Menu:  renderSharedListMenuEntries(m.menu),
	}
}
