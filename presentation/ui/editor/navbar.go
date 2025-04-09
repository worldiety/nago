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
)

type TNavbar struct {
	top     []core.View
	bottom  []core.View
	visible bool
}

func Navbar() TNavbar {
	return TNavbar{
		visible: true,
	}
}

func (c TNavbar) Top(views ...core.View) TNavbar {
	c.top = views
	return c
}

func (c TNavbar) Bottom(views ...core.View) TNavbar {
	c.bottom = views
	return c
}

func (c TNavbar) Render(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		ui.VStack(c.top...),
		ui.Spacer(),
		ui.VStack(c.bottom...),
	).Position(ui.Position{
		Type:   ui.PositionFixed,
		Left:   "0px",
		Top:    headerHeight,
		Bottom: "0px",
	}).Border(ui.Border{RightColor: ui.M7, RightWidth: ui.L1}).
		Frame(ui.Frame{Width: headerHeight}).
		Render(ctx)
}
