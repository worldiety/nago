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

type TScreen struct {
	title              string
	header             THeader
	navbar             TNavbar
	leadingToolWindows []TVToolWindow
	toolwindowTrailing TVToolWindow
	content            TContent
	modals             []core.View
}

func Screen(title string) TScreen {
	return TScreen{
		title: title,
	}
}

func (c TScreen) Header(header THeader) TScreen {
	c.header = header
	return c
}

func (c TScreen) Navbar(navbar TNavbar) TScreen {
	c.navbar = navbar
	return c
}

func (c TScreen) LeadingToolWindows(leading ...TVToolWindow) TScreen {
	c.leadingToolWindows = leading
	return c
}

func (c TScreen) TrailingToolWindow(tailing TVToolWindow) TScreen {
	c.toolwindowTrailing = tailing
	return c
}

func (c TScreen) Content(content TContent) TScreen {
	c.content = content
	return c
}

func (c TScreen) Modals(modals ...core.View) TScreen {
	c.modals = modals
	return c
}

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
