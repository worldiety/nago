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
	"go.wdy.de/nago/presentation/ui/editor"
)

// EditorContentStyle is the appearance of the content of an editor screen, like editor.ContentStyle of nago.
type EditorContentStyle int32

const (
	// EditorContentFull shows the content in a scrollable area, which fills the screen.
	EditorContentFull EditorContentStyle = 0
	// EditorContentPage shows the content as a styled page.
	EditorContentPage EditorContentStyle = 1
)

// TEditorContent is the main content of an editor screen, see EditorContent.
type TEditorContent struct{ ref js.Ref }

func (TEditorContent) isView() {}

// EditorContent returns the content, which shows the view and is visible by default, like editor.Content of nago.
//
//gostrict:host
func EditorContent(view View) TEditorContent {
	return TEditorContent{ref: wrap(editor.Content(viewOf(view)))}
}

// Style sets the appearance of the content.
//
//gostrict:host
func (c TEditorContent) Style(style EditorContentStyle) TEditorContent {
	return TEditorContent{ref: wrap(get[editor.TContent](c.ref).Style(editor.ContentStyle(style)))}
}

// TEditorHeader is the header of an editor screen, see EditorHeader.
type TEditorHeader struct{ ref js.Ref }

func (TEditorHeader) isView() {}

// EditorHeader returns the header of the window, which shows a menu with a back action by default, like
// editor.Header of nago.
//
//gostrict:host
func EditorHeader(wnd Window) TEditorHeader {
	return TEditorHeader{ref: wrap(editor.Header(wnd.nago()))}
}

// Center sets the views in the center of the header.
//
//gostrict:host
func (c TEditorHeader) Center(views ...View) TEditorHeader {
	return TEditorHeader{ref: wrap(get[editor.THeader](c.ref).Center(viewsOf(views)...))}
}

// Leading sets the views at the leading side of the header.
//
//gostrict:host
func (c TEditorHeader) Leading(views ...View) TEditorHeader {
	return TEditorHeader{ref: wrap(get[editor.THeader](c.ref).Leading(viewsOf(views)...))}
}

// TEditorNavbar is the navigation bar of an editor screen, see EditorNavbar.
type TEditorNavbar struct{ ref js.Ref }

func (TEditorNavbar) isView() {}

// EditorNavbar returns a navigation bar, which is visible by default, like editor.Navbar of nago.
//
//gostrict:host
func EditorNavbar() TEditorNavbar {
	return TEditorNavbar{ref: wrap(editor.Navbar())}
}

// Top sets the views at the top of the navigation bar.
//
//gostrict:host
func (c TEditorNavbar) Top(views ...View) TEditorNavbar {
	return TEditorNavbar{ref: wrap(get[editor.TNavbar](c.ref).Top(viewsOf(views)...))}
}

// Bottom sets the views at the bottom of the navigation bar.
//
//gostrict:host
func (c TEditorNavbar) Bottom(views ...View) TEditorNavbar {
	return TEditorNavbar{ref: wrap(get[editor.TNavbar](c.ref).Bottom(viewsOf(views)...))}
}

// TEditorVToolWindow is a tool window of an editor screen, e.g. a list at the side, see EditorToolWindow.
type TEditorVToolWindow struct{ ref js.Ref }

func (TEditorVToolWindow) isView() {}

// EditorToolWindow returns a tool window with the icon and the name, which is visible by default, like editor.ToolWindow
// of nago.
//
//gostrict:host
func EditorToolWindow(icon SVG, name string) TEditorVToolWindow {
	return TEditorVToolWindow{ref: wrap(editor.ToolWindow(core.SVG(icon), name))}
}

// Visible shows or hides the tool window.
//
//gostrict:host
func (c TEditorVToolWindow) Visible(visible bool) TEditorVToolWindow {
	return TEditorVToolWindow{ref: wrap(get[editor.TVToolWindow](c.ref).Visible(visible))}
}

// Top sets the view at the top of the tool window. A view, which is not decorated, is ignored.
//
//gostrict:host
func (c TEditorVToolWindow) Top(v View) TEditorVToolWindow {
	return TEditorVToolWindow{ref: wrap(get[editor.TVToolWindow](c.ref).Top(c6Decored(v)))}
}

// Bottom sets the view at the bottom of the tool window. A view, which is not decorated, is ignored.
//
//gostrict:host
func (c TEditorVToolWindow) Bottom(v View) TEditorVToolWindow {
	return TEditorVToolWindow{ref: wrap(get[editor.TVToolWindow](c.ref).Bottom(c6Decored(v)))}
}

// Content sets the main content of the tool window. A view, which is not decorated, is ignored.
//
//gostrict:host
func (c TEditorVToolWindow) Content(v View) TEditorVToolWindow {
	return TEditorVToolWindow{ref: wrap(get[editor.TVToolWindow](c.ref).Content(c6Decored(v)))}
}

// TEditorScreen is a full application screen, see EditorScreen.
type TEditorScreen struct{ ref js.Ref }

func (TEditorScreen) isView() {}

// EditorScreen returns a screen with the title, like editor.Screen of nago.
//
//gostrict:host
func EditorScreen(title string) TEditorScreen {
	return TEditorScreen{ref: wrap(editor.Screen(title))}
}

// Header sets the header of the screen.
//
//gostrict:host
func (c TEditorScreen) Header(header TEditorHeader) TEditorScreen {
	return TEditorScreen{ref: wrap(get[editor.TScreen](c.ref).Header(get[editor.THeader](header.ref)))}
}

// Navbar sets the navigation bar of the screen.
//
//gostrict:host
func (c TEditorScreen) Navbar(navbar TEditorNavbar) TEditorScreen {
	return TEditorScreen{ref: wrap(get[editor.TScreen](c.ref).Navbar(get[editor.TNavbar](navbar.ref)))}
}

// LeadingToolWindows sets the tool windows at the leading side of the screen.
//
//gostrict:host
func (c TEditorScreen) LeadingToolWindows(leading ...TEditorVToolWindow) TEditorScreen {
	tws := make([]editor.TVToolWindow, 0, len(leading))
	for _, tw := range leading {
		tws = append(tws, get[editor.TVToolWindow](tw.ref))
	}
	return TEditorScreen{ref: wrap(get[editor.TScreen](c.ref).LeadingToolWindows(tws...))}
}

// TrailingToolWindow sets the tool window at the trailing side of the screen.
//
//gostrict:host
func (c TEditorScreen) TrailingToolWindow(tailing TEditorVToolWindow) TEditorScreen {
	return TEditorScreen{ref: wrap(get[editor.TScreen](c.ref).TrailingToolWindow(get[editor.TVToolWindow](tailing.ref)))}
}

// Content sets the main content area of the screen.
//
//gostrict:host
func (c TEditorScreen) Content(content TEditorContent) TEditorScreen {
	return TEditorScreen{ref: wrap(get[editor.TScreen](c.ref).Content(get[editor.TContent](content.ref)))}
}

// Modals sets the modal dialogs of the screen.
//
//gostrict:host
func (c TEditorScreen) Modals(modals ...View) TEditorScreen {
	return TEditorScreen{ref: wrap(get[editor.TScreen](c.ref).Modals(viewsOf(modals)...))}
}
