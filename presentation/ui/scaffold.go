package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type ScaffoldAlignment string

func (s ScaffoldAlignment) ora() ora.ScaffoldAlignment {
	return ora.ScaffoldAlignment(s)
}

const (
	ScaffoldAlignmentTop     ScaffoldAlignment = "u"
	ScaffoldAlignmentLeading ScaffoldAlignment = "l"
)

// ScaffoldMenuEntry represents either a menu node or leaf. See also helper functions [ForwardScaffoldMenuEntry] and
// [ParentScaffoldMenuEntry].
type ScaffoldMenuEntry struct {
	Icon core.View
	// IconActive is optional
	IconActive core.View
	Title      string
	Action     func()
	// MarkAsActiveAt contains the factory id at which this entry shall be highlighted automatically as active.
	// Use . for index.
	MarkAsActiveAt core.NavigationPath
	Menu           []ScaffoldMenuEntry

	// intentionally left out expanded and badge, because badge can be emulated with Box layout and expanded is automatic
}

func ForwardScaffoldMenuEntry(wnd core.Window, icon core.SVG, title string, dst core.NavigationPath) ScaffoldMenuEntry {
	return ScaffoldMenuEntry{
		Icon:  Image().Embed(icon).Frame(Frame{}.Size(L24, L24)),
		Title: title,
		Action: func() {
			wnd.Navigation().ForwardTo(dst, nil)
		},
		MarkAsActiveAt: dst,
	}
}

func ParentScaffoldMenuEntry(wnd core.Window, icon core.SVG, title string, children ...ScaffoldMenuEntry) ScaffoldMenuEntry {
	return ScaffoldMenuEntry{
		Icon:  Image().Embed(icon).Frame(Frame{}.Size(L24, L24)),
		Title: title,
		Menu:  children,
	}
}

type TScaffold struct {
	logo      core.View
	body      core.View
	alignment ora.ScaffoldAlignment
	menu      []ScaffoldMenuEntry
}

func Scaffold(alignment ScaffoldAlignment) TScaffold {
	return TScaffold{alignment: alignment.ora()}
}

func (c TScaffold) Logo(view core.View) TScaffold {
	c.logo = view
	return c
}

func (c TScaffold) Body(view core.View) TScaffold {
	c.body = view
	return c
}

func (c TScaffold) Menu(items ...ScaffoldMenuEntry) TScaffold {
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

func makeMenu(ctx core.RenderContext, menu []ScaffoldMenuEntry) []ora.ScaffoldMenuEntry {
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
			Factory:    ora.ComponentFactoryId(entry.MarkAsActiveAt),
			Menu:       makeMenu(ctx, entry.Menu),
		})
	}

	return menuEntries
}
