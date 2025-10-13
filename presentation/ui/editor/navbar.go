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

// TNavbar is a composite component (Navbar).
// This component organizes views into top and bottom sections
// within a vertical navigation bar, which can be shown or hidden.
type TNavbar struct {
	top     []core.View // views displayed at the top of the navbar
	bottom  []core.View // views displayed at the bottom of the navbar
	visible bool        // controls whether the navbar is visible
}

// Navbar creates a new TNavbar with visibility enabled by default.
func Navbar() TNavbar {
	return TNavbar{
		visible: true,
	}
}

// Top sets the top section of the navbar to the given views.
func (c TNavbar) Top(views ...core.View) TNavbar {
	c.top = views
	return c
}

// Bottom sets the bottom section of the navbar to the given views.
func (c TNavbar) Bottom(views ...core.View) TNavbar {
	c.bottom = views
	return c
}

// Render builds and returns the RenderNode for the TNavbar.
// It stacks the top views, a spacer, and the bottom views vertically,
// positions the navbar fixed to the left side below the header,
// and applies a border and fixed width.
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
