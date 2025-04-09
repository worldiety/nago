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

type THeader struct {
	leadingCorner core.View
	leading       core.View
	center        core.View
	trailing      core.View
	visible       bool
}

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

func (c THeader) Center(views ...core.View) THeader {
	c.center = ui.VStack(views...)
	return c
}

func (c THeader) Leading(views ...core.View) THeader {
	c.leading = ui.VStack(views...)
	return c
}

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
