// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tags

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// TColoredTextPill is a basic component (Colored Text Pill).
// It displays a short text label inside a pill-shaped container,
// styled with a background color, padding, and rounded borders.
// Pills are often used to represent tags, statuses, or categories.
type TColoredTextPill struct {
	color ui.Color // background color of the pill
	text  string   // text displayed inside the pill

	padding            ui.Padding // spacing around the text inside the pill
	frame              ui.Frame   // layout frame for sizing and alignment
	border             ui.Border  // border styling, typically rounded
	accessibilityLabel string     // accessibility label for screen readers
	invisible          bool       // when true, the pill is not rendered
}

// ColoredTextPill creates a new pill with the given background color and text,
// applying default padding and rounded borders.
func ColoredTextPill(color ui.Color, text string) TColoredTextPill {
	return TColoredTextPill{
		color:   color,
		text:    text,
		padding: ui.Padding{}.Horizontal(ui.L8).Vertical(ui.L4),
		border:  ui.Border{}.Radius(ui.L16),
	}
}

// Padding sets the inner spacing around the pill's text.
func (t TColoredTextPill) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

// WithFrame applies a transformation function to the pill's frame
// and returns the updated component.
func (t TColoredTextPill) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

// Frame sets the layout frame of the pill, including size and alignment.
func (t TColoredTextPill) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

// Border sets the border style of the pill, such as radius or thickness.
func (t TColoredTextPill) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

// Visible controls the visibility of the pill; setting false hides it.
func (t TColoredTextPill) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

// AccessibilityLabel sets a label used by screen readers for accessibility.
func (t TColoredTextPill) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}

func (t TColoredTextPill) Render(ctx core.RenderContext) core.RenderNode {
	return ui.HStack(
		ui.Text(t.text).Color(ui.ColorBlack),
	).BackgroundColor(t.color).
		Padding(t.padding).
		Border(t.border).
		Render(ctx)
}
