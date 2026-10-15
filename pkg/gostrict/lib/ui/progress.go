// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/progress"
)

// ProgressStyle is the visual style of a progress bar, like progress.Style of nago.
type ProgressStyle int32

const (
	ProgressHorizontal ProgressStyle = 0
)

// TProgress is a progress bar, which shows the completion state of a task as a filled bar, see progress.TProgress of
// nago.
type TProgress struct{ ref js.Ref }

func (TProgress) isView() {}

// ProgressLinearProgress returns a horizontal progress bar with the default accent color, full width and the standard
// height, like progress.LinearProgress of nago.
//
//gostrict:host
func ProgressLinearProgress() TProgress {
	return TProgress{ref: wrap(progress.LinearProgress())}
}

// Style sets the visual style of the progress bar.
//
//gostrict:host
func (c TProgress) Style(style ProgressStyle) TProgress {
	return TProgress{ref: wrap(get[progress.TProgress](c.ref).Style(progress.Style(style)))}
}

// Color sets the foreground color of the progress indicator.
//
//gostrict:host
func (c TProgress) Color(color Color) TProgress {
	return TProgress{ref: wrap(get[progress.TProgress](c.ref).Color(nui.Color(color)))}
}

// BackgroundColor sets the background color of the unfilled portion of the bar.
//
//gostrict:host
func (c TProgress) BackgroundColor(color Color) TProgress {
	return TProgress{ref: wrap(get[progress.TProgress](c.ref).BackgroundColor(nui.Color(color)))}
}

// FullWidth sets the progress bar to span the full available width.
//
//gostrict:host
func (c TProgress) FullWidth() TProgress {
	return TProgress{ref: wrap(get[progress.TProgress](c.ref).FullWidth())}
}

// Progress sets the progress from 0 to 1, values outside are clamped.
//
//gostrict:host
func (c TProgress) Progress(v float64) TProgress {
	return TProgress{ref: wrap(get[progress.TProgress](c.ref).Progress(v))}
}

// Frame sets the layout frame of the progress bar, including size and spacing.
//
//gostrict:host
func (c TProgress) Frame(frame Frame) TProgress {
	return TProgress{ref: wrap(get[progress.TProgress](c.ref).Frame(frame.nago()))}
}
