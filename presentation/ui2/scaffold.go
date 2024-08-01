package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type MenuEntry struct {
	Icon       core.View
	IconActive core.View
	Title      string
	Action     func()
	Factory    ora.ComponentFactoryId
	Menu       []MenuEntry
	Badge      string
	Expanded   bool
}

type TScaffold struct {
	logo      core.View
	body      core.View
	alignment ora.ScaffoldAlignment
	menu      []MenuEntry
}

func Scaffold(alignment ora.ScaffoldAlignment) TScaffold {
	return TScaffold{alignment: alignment}
}

func (c TScaffold) Logo(view core.View) TScaffold {
	c.logo = view
	return c
}

func (c TScaffold) Body(view core.View) TScaffold {
	c.body = view
	return c
}

func (c TScaffold) Menu(items ...MenuEntry) TScaffold {
	c.menu = items
	return c
}

func (c TScaffold) Render(ctx core.RenderContext) ora.Component {

	return ora.Scaffold{
		Type:      ora.ScaffoldT,
		Body:      render(ctx, c.body),
		Logo:      render(ctx, c.logo),
		Menu:      makeMenu(ctx, c.menu),
		Alignment: c.alignment,
	}
}

func makeMenu(ctx core.RenderContext, menu []MenuEntry) []ora.ScaffoldMenuEntry {
	if len(menu) == 0 {
		return nil
	}

	menuEntries := make([]ora.ScaffoldMenuEntry, 0, len(menu))
	for _, entry := range menu {
		menuEntries = append(menuEntries, ora.ScaffoldMenuEntry{
			Icon:       render(ctx, entry.Icon),
			IconActive: render(ctx, entry.IconActive),
			Title:      entry.Title,
			Action:     ctx.MountCallback(entry.Action),
			Factory:    entry.Factory,
			Menu:       makeMenu(ctx, entry.Menu),
			Badge:      entry.Badge,
			Expanded:   entry.Expanded,
		})
	}

	return menuEntries
}
