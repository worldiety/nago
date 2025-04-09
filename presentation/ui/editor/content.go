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

type ContentStyle int

const (
	ContentFull ContentStyle = iota
	ContentPage
)

type TContent struct {
	view    core.View
	visible bool
	style   ContentStyle
}

func Content(view core.View) TContent {
	return TContent{
		view:    view,
		visible: true,
	}
}

func (c TContent) Style(style ContentStyle) TContent {
	c.style = style
	return c
}

func (c TContent) Render(ctx core.RenderContext) core.RenderNode {
	if c.style == ContentFull {
		return ui.ScrollView(c.view).
			Axis(ui.ScrollViewAxisBoth).
			Position(ui.Position{
				Type:   ui.PositionFixed,
				Left:   headerHeight,
				Top:    headerHeight,
				Right:  "0px",
				Bottom: "0px",
			}).Render(ctx)
	}

	if c.style == ContentPage {
		return ui.VStack(
			ui.ScrollView(c.view).
				Axis(ui.ScrollViewAxisBoth).
				Border(ui.Border{}.Shadow(ui.L4)).
				Padding(ui.Padding{}.All(ui.L32)).
				BackgroundColor(ui.M1).
				Frame(ui.Frame{MaxWidth: "100rem", Width: ui.Full, MinHeight: ui.Full}),
		).Position(ui.Position{
			Type:   ui.PositionFixed,
			Left:   headerHeight,
			Top:    headerHeight,
			Right:  "0px",
			Bottom: "0px",
		}).Render(ctx)
	}

	return ui.Text("invalid content style").Render(ctx)
}
