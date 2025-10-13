// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package editor

import (
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

// THeader is a composite container (Header).
// This component organizes views into distinct regions (leading corner,
// leading, center, and trailing) typically used in application headers.
// It can also be toggled visible or hidden.
type THeader struct {
	leadingCorner core.View // optional element placed in the leading corner (e.g., logo or icon)
	leading       core.View // element aligned at the leading side (e.g., navigation or menu)
	center        core.View // main element centered in the header (e.g., title or search bar)
	trailing      core.View // element aligned at the trailing side (e.g., actions or profile menu)
	visible       bool      // controls whether the header is visible
}

// Header creates a new THeader for the given window.
// By default, it is visible and includes a menu button in the leading corner
// with a "Back" navigation item.
func Header(wnd core.Window) THeader {
	return THeader{
		visible: true,
		leadingCorner: ui.Menu(ui.TertiaryButton(nil).PreIcon(flowbiteOutline.Bars), ui.MenuGroup(
			ui.MenuItem(func() {
				wnd.Navigation().Back()
			}, ui.Text("Zurück")),
		)),
	}
}

// Center sets the center region of the header to the given views.
func (c THeader) Center(views ...core.View) THeader {
	c.center = ui.VStack(views...)
	return c
}

// Leading sets the leading region of the header to the given views.
func (c THeader) Leading(views ...core.View) THeader {
	c.leading = ui.VStack(views...)
	return c
}

// Render builds and returns the RenderNode for the THeader.
// It arranges the header into leading corner, leading, center, and trailing regions
// inside a horizontal stack, separated by spacers and borders, and positions it
// fixed at the top of the window.
func (c THeader) Render(ctx core.RenderContext) core.RenderNode {
	return ui.HStack(
		ui.HStack(c.leadingCorner).Border(ui.Border{RightColor: ui.M7, RightWidth: ui.L1}).Frame(ui.Frame{Height: headerHeight, Width: headerHeight}),
		ui.IfFunc(c.leading != nil, func() core.View {
			return ui.HStack(c.leading, ui.VLine().Frame(ui.Frame{Width: ui.L1})).Alignment(ui.Stretch)
		}),
		ui.Spacer(),
		c.center,
		ui.Spacer(),
		c.trailing,
	).Position(ui.Position{
		Type:  ui.PositionFixed,
		Left:  "0px",
		Top:   "0px",
		Right: "0px",
	}).
		Border(ui.Border{BottomColor: ui.M7, BottomWidth: ui.L1}).
		Frame(ui.Frame{Height: headerHeight}).
		Render(ctx)
}
