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

// ForwardScaffoldMenuEntry creates a leaf menu entry that navigates
// forward to the given destination when clicked. It also sets up
// automatic active-state highlighting based on the destination path.
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

// ParentScaffoldMenuEntry creates a parent menu entry that groups together
// multiple child entries. Unlike ForwardScaffoldMenuEntry, this entry does not
// navigate directly, but instead serves as a container for its children.
func ParentScaffoldMenuEntry(wnd core.Window, icon core.SVG, title string, children ...ScaffoldMenuEntry) ScaffoldMenuEntry {
	return ScaffoldMenuEntry{
		Icon:  Image().Embed(icon).Frame(Frame{}.Size(L20, L20)),
		Title: title,
		Menu:  children,
	}
}

// TScaffold is a composite component (Scaffold).
// It provides a structural layout for applications with a consistent navigation
// and content framework. A scaffold usually consists of a logo, navigation menu,
// main body, footer, and optional bottom view. It also supports responsive design
// through alignment and breakpoints.
type TScaffold struct {
	logo       core.View               // logo or brand element
	body       core.View               // main content body
	alignment  proto.ScaffoldAlignment // scaffold alignment (e.g., left, right, top)
	menu       []ScaffoldMenuEntry     // navigation menu entries
	bottomView core.View               // optional bottom view (e.g., settings, profile)
	breakpoint int                     // breakpoint for responsive layout
	footer     core.View               // footer content
}

// Scaffold creates a new scaffold with the given alignment.
func Scaffold(alignment ScaffoldAlignment) TScaffold {
	return TScaffold{alignment: alignment.ora()}
}

// Logo sets the logo or brand element of the scaffold.
func (c TScaffold) Logo(view core.View) TScaffold {
	c.logo = view
	return c
}

// Body sets the main content body of the scaffold.
func (c TScaffold) Body(view core.View) TScaffold {
	c.body = view
	return c
}

// Menu sets the navigation menu entries for the scaffold.
func (c TScaffold) Menu(items ...ScaffoldMenuEntry) TScaffold {
	c.menu = items
	return c
}

// BottomView sets the optional bottom view of the scaffold,
// often used for secondary actions like user profile or settings.
func (c TScaffold) BottomView(view core.View) TScaffold {
	c.bottomView = view
	return c
}

// Footer sets the footer content of the scaffold.
func (c TScaffold) Footer(view core.View) TScaffold {
	c.footer = view
	return c
}

// Breakpoint sets the responsive breakpoint at which the scaffold layout
// may adapt (e.g., switch between drawer and permanent menu).
func (c TScaffold) Breakpoint(breakpoint int) TScaffold {
	c.breakpoint = breakpoint
	return c
}

// Render builds and returns the protocol representation of the scaffold.
func (c TScaffold) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.Scaffold{
		Body:       render(ctx, c.body),
		Logo:       render(ctx, c.logo),
		Menu:       makeMenu(ctx, c.menu),
		BottomView: render(ctx, c.bottomView),
		Alignment:  c.alignment,
		Breakpoint: proto.Uint(c.breakpoint),
		Footer:     render(ctx, c.footer),
	}
}

// makeMenu recursively converts a list of ScaffoldMenuEntry objects
// into their protocol representation. It renders icons, titles, actions,
// active state markers, and nested sub-menus. Returns nil if no entries exist.
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
