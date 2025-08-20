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

// TContent is a composite component (Content).
// This component holds a single view with an associated style
// and can be toggled visible or hidden.
type TContent struct {
	view    core.View    // the inner view displayed within the content
	visible bool         // controls whether the content is visible
	style   ContentStyle // defines the appearance of the content
}

// Content creates a new TContent with the given view and sets it visible by default.
func Content(view core.View) TContent {
	return TContent{
		view:    view,
		visible: true,
	}
}

// Style applies the given ContentStyle to the TContent.
func (c TContent) Style(style ContentStyle) TContent {
	c.style = style
	return c
}

// Render builds and returns the RenderNode for the TContent.
// It applies layout and styling depending on the ContentStyle:
// - ContentFull: renders the view in a scrollable full-area container.
// - ContentPage: renders the view in a scrollable, styled page container.
// Returns a text node if the style is invalid.
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
