// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// ScaffoldAlignment is the alignment of a scaffold, see Scaffold.
type ScaffoldAlignment uint32

const (
	// ScaffoldAlignmentTop aligns the scaffold at the top.
	ScaffoldAlignmentTop ScaffoldAlignment = 0
	// ScaffoldAlignmentLeading aligns the scaffold at the leading edge.
	ScaffoldAlignmentLeading ScaffoldAlignment = 1
)

// NavigationPath is the path of a root view, to which a navigation goes, e.g. "a/b/c".
type NavigationPath string

// ScaffoldMenuEntry is an entry of the menu of a scaffold, either a leaf or a parent, see ForwardScaffoldMenuEntry
// and ParentScaffoldMenuEntry. It is a building block, which the menu of a scaffold takes.
type ScaffoldMenuEntry struct{ ref js.Ref }

// TScaffold is a scaffold, see Scaffold. It lays out an application with a navigation menu, a logo, a body, a footer
// and an optional bottom view.
type TScaffold struct{ ref js.Ref }

func (TScaffold) isView() {}

// DefaultScaffoldBreakpoint is the window width, from which a scaffold shows its navigation bar or sidebar instead of
// the burger menu, unless Breakpoint sets another one. It is measured like the width of the window.
const DefaultScaffoldBreakpoint int32 = 768

// Scaffold returns a scaffold with the alignment.
//
//gostrict:host
func Scaffold(alignment ScaffoldAlignment) TScaffold {
	return TScaffold{ref: wrap(nui.Scaffold(nui.ScaffoldAlignment(alignment)))}
}

// Alignment sets the alignment of the scaffold.
//
//gostrict:host
func (c TScaffold) Alignment(alignment ScaffoldAlignment) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).Alignment(nui.ScaffoldAlignment(alignment)))}
}

// Logo sets the logo or brand element of the scaffold.
//
//gostrict:host
func (c TScaffold) Logo(view View) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).Logo(viewOf(view)))}
}

// Body sets the main content body of the scaffold.
//
//gostrict:host
func (c TScaffold) Body(view View) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).Body(viewOf(view)))}
}

// Menu sets the navigation menu entries of the scaffold.
//
//gostrict:host
func (c TScaffold) Menu(items ...ScaffoldMenuEntry) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).Menu(b5MenuEntries(items)...))}
}

// BottomView sets the optional bottom view of the scaffold, often used for secondary actions like the user profile or
// the settings.
//
//gostrict:host
func (c TScaffold) BottomView(view View) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).BottomView(viewOf(view)))}
}

// Footer sets the footer content of the scaffold.
//
//gostrict:host
func (c TScaffold) Footer(view View) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).Footer(viewOf(view)))}
}

// Breakpoint sets the window width, from which the scaffold shows its navigation bar or sidebar instead of the burger
// menu. Zero means DefaultScaffoldBreakpoint. See also ScaffoldMenuBarVisible.
//
//gostrict:host
func (c TScaffold) Breakpoint(breakpoint int32) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).Breakpoint(int(breakpoint)))}
}

// Height sets the fixed height of the scaffold.
//
//gostrict:host
func (c TScaffold) Height(height Length) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).Height(nui.Length(height)))}
}

// BodyFullSize lets the body use the full size of the screen.
//
//gostrict:host
func (c TScaffold) BodyFullSize(fullSize bool) TScaffold {
	return TScaffold{ref: wrap(get[nui.TScaffold](c.ref).BodyFullSize(fullSize))}
}

// ForwardScaffoldMenuEntry returns a leaf entry of the menu, which navigates forward to dst, when the user clicks it.
// The entry is highlighted as active, while the window shows dst.
//
//gostrict:host
func ForwardScaffoldMenuEntry(wnd Window, icon SVG, title string, dst NavigationPath) ScaffoldMenuEntry {
	return ScaffoldMenuEntry{ref: wrap(nui.ForwardScaffoldMenuEntry(wnd.nago(), core.SVG(icon), title, core.NavigationPath(dst)))}
}

// ParentScaffoldMenuEntry returns a parent entry of the menu, which groups the children. Unlike
// ForwardScaffoldMenuEntry, it does not navigate itself.
//
//gostrict:host
func ParentScaffoldMenuEntry(wnd Window, icon SVG, title string, children ...ScaffoldMenuEntry) ScaffoldMenuEntry {
	return ScaffoldMenuEntry{ref: wrap(nui.ParentScaffoldMenuEntry(wnd.nago(), core.SVG(icon), title, b5MenuEntries(children)...))}
}

// ScaffoldMenuBarVisible reports whether the scaffold around the current page shows its navigation bar or sidebar. It
// is false, if the scaffold shows the burger menu. Use it to adapt a page to the space, which the menu takes.
//
//gostrict:host
func ScaffoldMenuBarVisible(wnd Window) bool {
	return nui.ScaffoldMenuBarVisible(wnd.nago())
}
