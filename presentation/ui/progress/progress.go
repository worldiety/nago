// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package progress

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type Style int

const (
	Horizontal Style = iota
)

type TProgress struct {
	style           Style
	color           ui.Color
	backgroundColor ui.Color
	progress        float64
	frame           ui.Frame
}

func LinearProgress() TProgress {
	return TProgress{
		style:           Horizontal,
		color:           ui.ColorAccent,
		backgroundColor: ui.ColorCardFooter,
		frame: ui.Frame{
			Height: ui.L4,
			Width:  ui.Full,
		},
	}
}

func (c TProgress) Style(style Style) TProgress {
	c.style = style
	return c
}

func (c TProgress) Color(color ui.Color) TProgress {
	c.color = color
	return c
}

func (c TProgress) BackgroundColor(color ui.Color) TProgress {
	c.backgroundColor = color
	return c
}

func (c TProgress) FullWidth() TProgress {
	c.frame.Width = ui.Full
	return c
}

// Progress must be between 0 and 1. Values are clamped.
func (c TProgress) Progress(v float64) TProgress {
	c.progress = max(min(v, 1), 0)
	return c
}

func (c TProgress) Frame(frame ui.Frame) TProgress {
	c.frame = frame
	return c
}

func (c TProgress) Render(ctx core.RenderContext) core.RenderNode {
	return ui.HStack(
		ui.HStack().BackgroundColor(c.color).Border(ui.Border{}.Circle()).Frame(ui.Frame{Height: c.frame.Height, Width: ui.Relative(core.Weight(c.progress))}),
	).Alignment(ui.Leading).BackgroundColor(c.backgroundColor).Frame(c.frame).Border(ui.Border{}.Circle()).Render(ctx)
}
