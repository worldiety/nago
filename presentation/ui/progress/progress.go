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

// TProgress is a composite component (Progress).
// It represents the completion state of a task or process as a filled bar.
// The style, color, and background can be customized, and the value is given
// as a floating-point percentage between 0.0 and 1.0.
type TProgress struct {
	style           Style    // visual style variant of the progress bar
	color           ui.Color // foreground color of the progress indicator
	backgroundColor ui.Color // background color of the unfilled portion
	progress        float64  // progress value, from 0.0 (empty) to 1.0 (full)
	frame           ui.Frame // layout container / sizing and spacing context
}

// LinearProgress creates a horizontal progress bar with default accent color,
// card footer background, full width, and standard height.
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

// Style sets the visual style of the progress bar (e.g., horizontal or circular).
func (c TProgress) Style(style Style) TProgress {
	c.style = style
	return c
}

// Color sets the foreground color of the progress indicator.
func (c TProgress) Color(color ui.Color) TProgress {
	c.color = color
	return c
}

// BackgroundColor sets the background color of the unfilled portion of the bar.
func (c TProgress) BackgroundColor(color ui.Color) TProgress {
	c.backgroundColor = color
	return c
}

// FullWidth sets the progress bar to span the full available width.
func (c TProgress) FullWidth() TProgress {
	c.frame.Width = ui.Full
	return c
}

// Progress must be between 0 and 1. Values are clamped.
func (c TProgress) Progress(v float64) TProgress {
	c.progress = max(min(v, 1), 0)
	return c
}

// Frame sets the layout frame of the progress bar, including size and spacing.
func (c TProgress) Frame(frame ui.Frame) TProgress {
	c.frame = frame
	return c
}

// Render builds and returns the visual representation of the progress bar.
// It draws a filled portion proportional to the progress value inside a
// background container, applying colors, borders, and layout settings.
func (c TProgress) Render(ctx core.RenderContext) core.RenderNode {
	return ui.HStack(
		ui.HStack().BackgroundColor(c.color).Border(ui.Border{}.Circle()).Frame(ui.Frame{Height: c.frame.Height, Width: ui.Relative(core.Weight(c.progress))}),
	).Alignment(ui.Leading).BackgroundColor(c.backgroundColor).Frame(c.frame).Border(ui.Border{}.Circle()).Render(ctx)
}
