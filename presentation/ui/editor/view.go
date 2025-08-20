// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package editor

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

const (
	headerHeight = ui.L40
	toolbarWidth = ui.L320
)

// TScreen represents a container (Screen).
// This component models a full application screen with title, header,
// navbar, tool windows, main content, and optional modals.
type TScreen struct {
	title              string
	header             THeader
	navbar             TNavbar
	leadingToolWindows []TVToolWindow
	toolwindowTrailing TVToolWindow
	content            TContent
	modals             []core.View
}

// Screen creates a new TScreen with the given title.
func Screen(title string) TScreen {
	return TScreen{
		title: title,
	}
}

// Header sets the header section of the screen.
func (c TScreen) Header(header THeader) TScreen {
	c.header = header
	return c
}

// Navbar sets the navigation bar of the screen.
func (c TScreen) Navbar(navbar TNavbar) TScreen {
	c.navbar = navbar
	return c
}

// LeadingToolWindows sets the leading-side tool windows of the screen.
func (c TScreen) LeadingToolWindows(leading ...TVToolWindow) TScreen {
	c.leadingToolWindows = leading
	return c
}

// TrailingToolWindow sets the trailing-side tool window of the screen.
func (c TScreen) TrailingToolWindow(tailing TVToolWindow) TScreen {
	c.toolwindowTrailing = tailing
	return c
}

// Content sets the main content area of the screen.
func (c TScreen) Content(content TContent) TScreen {
	c.content = content
	return c
}

// Modals sets the modal dialogs of the screen.
func (c TScreen) Modals(modals ...core.View) TScreen {
	c.modals = modals
	return c
}

// Render builds and returns the RenderNode for the TScreen.
// The entire screen is rendered as a fixed, full-size container with
// a background color.
func (c TScreen) Render(ctx core.RenderContext) core.RenderNode {
	var tmp []core.View
	tmp = append(tmp,
		ui.WindowTitle(c.title),
		alert.BannerMessages(ctx.Window()),
		ui.IfFunc(c.navbar.visible, func() core.View {
			return c.navbar
		}),

		ui.IfFunc(c.content.visible, func() core.View {
			return c.content
		}),
	)

	for _, tw := range c.leadingToolWindows {
		if tw.visible {
			tmp = append(tmp, ui.VStack(
				tw,
			).
				Position(ui.Position{
					Type:   ui.PositionFixed,
					Left:   headerHeight,
					Top:    headerHeight,
					Bottom: "0px",
				}).
				BackgroundColor(ui.M6).
				Border(ui.Border{}.Shadow(ui.L4)).
				Frame(ui.Frame{Width: toolbarWidth}),
			)
		}
	}
	tmp = append(tmp,
		ui.IfFunc(c.toolwindowTrailing.visible, func() core.View {
			return ui.VStack(
				c.toolwindowTrailing,
			).
				Position(ui.Position{
					Type:   ui.PositionFixed,
					Right:  "0px",
					Top:    headerHeight,
					Bottom: "0px",
				}).
				BackgroundColor(ui.M6).
				Border(ui.Border{}.Shadow(ui.L4)).
				Frame(ui.Frame{Width: toolbarWidth})
		}),

		ui.IfFunc(c.header.visible, func() core.View {
			return c.header
		}))

	tmp = append(tmp, c.modals...)

	return ui.VStack(
		tmp...,
	).Position(ui.Position{
		Type:   ui.PositionFixed,
		Left:   "0px",
		Top:    "0px",
		Right:  "0px",
		Bottom: "0px",
	}).
		BackgroundColor(ui.M4).
		Render(ctx)
}
