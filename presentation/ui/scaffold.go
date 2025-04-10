// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type ScaffoldAlignment uint

func (s ScaffoldAlignment) ora() proto.ScaffoldAlignment {
	return proto.ScaffoldAlignment(s)
}

const (
	ScaffoldAlignmentTop     ScaffoldAlignment = ScaffoldAlignment(proto.ScaffoldAlignmentTop)
	ScaffoldAlignmentLeading ScaffoldAlignment = ScaffoldAlignment(proto.ScaffoldAlignmentLeading)
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
		Icon:  Image().Embed(icon).Frame(Frame{}.Size(L20, L20)),
		Title: title,
		Action: func() {
			wnd.Navigation().ForwardTo(dst, nil)
		},
		MarkAsActiveAt: dst,
	}
}

func ParentScaffoldMenuEntry(wnd core.Window, icon core.SVG, title string, children ...ScaffoldMenuEntry) ScaffoldMenuEntry {
	return ScaffoldMenuEntry{
		Icon:  Image().Embed(icon).Frame(Frame{}.Size(L20, L20)),
		Title: title,
		Menu:  children,
	}
}

type TScaffold struct {
	logo       core.View
	body       core.View
	alignment  proto.ScaffoldAlignment
	menu       []ScaffoldMenuEntry
	breakpoint int
	footer     core.View
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

func (c TScaffold) Footer(view core.View) TScaffold {
	c.footer = view
	return c
}

func (c TScaffold) Breakpoint(breakpoint int) TScaffold {
	c.breakpoint = breakpoint
	return c
}

func (c TScaffold) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.Scaffold{
		Body:       render(ctx, c.body),
		Logo:       render(ctx, c.logo),
		Menu:       makeMenu(ctx, c.menu),
		Alignment:  c.alignment,
		Breakpoint: proto.Uint(c.breakpoint),
		Footer:     render(ctx, c.footer),
	}
}

func makeMenu(ctx core.RenderContext, menu []ScaffoldMenuEntry) []proto.ScaffoldMenuEntry {
	if len(menu) == 0 {
		return nil
	}

	menuEntries := make([]proto.ScaffoldMenuEntry, 0, len(menu))
	for _, entry := range menu {
		menuEntries = append(menuEntries, proto.ScaffoldMenuEntry{
			Icon:       render(ctx, entry.Icon),
			IconActive: render(ctx, entry.IconActive),
			Title:      proto.Str(entry.Title),
			Action:     ctx.MountCallback(entry.Action),
			RootView:   proto.RootViewID(entry.MarkAsActiveAt),
			Menu:       makeMenu(ctx, entry.Menu),
		})
	}

	return menuEntries
}
