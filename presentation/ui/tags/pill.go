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

// TColoredTextPill is a basic component(Colored Text Pill).
type TColoredTextPill struct {
	color ui.Color
	text  string

	padding            ui.Padding
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

func ColoredTextPill(color ui.Color, text string) TColoredTextPill {
	return TColoredTextPill{
		color:   color,
		text:    text,
		padding: ui.Padding{}.Horizontal(ui.L8).Vertical(ui.L4),
		border:  ui.Border{}.Radius(ui.L16),
	}
}

func (t TColoredTextPill) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TColoredTextPill) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TColoredTextPill) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TColoredTextPill) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TColoredTextPill) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

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
