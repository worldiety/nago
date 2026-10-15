// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import nui "go.wdy.de/nago/presentation/ui"

// Frame is the size of a view.
type Frame struct {
	MinWidth          Length
	MaxWidth          Length
	MinHeight         Length
	MaxHeight         Length
	Width             Length
	Height            Length
	FlexGrow          bool
	FlexPreventShrink bool
}

// Size sets the width and the height.
//
//gostrict:host
func (f Frame) Size(w, h Length) Frame {
	return frameFrom(f.nago().Size(nui.Length(w), nui.Length(h)))
}

// MatchScreen sets the frame to the full height and width of the viewport.
//
//gostrict:host
func (f Frame) MatchScreen() Frame {
	return frameFrom(f.nago().MatchScreen())
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (f Frame) FullWidth() Frame {
	return frameFrom(f.nago().FullWidth())
}

// FullHeight sets the height to the full height of the parent.
//
//gostrict:host
func (f Frame) FullHeight() Frame {
	return frameFrom(f.nago().FullHeight())
}

// Large sets the maximal width to L560 and the width to Full.
//
//gostrict:host
func (f Frame) Large() Frame {
	return frameFrom(f.nago().Large())
}

// Larger sets the maximal width to L880 and the width to Full.
//
//gostrict:host
func (f Frame) Larger() Frame {
	return frameFrom(f.nago().Larger())
}

// Grow lets the view grow in a stack.
//
//gostrict:host
func (f Frame) Grow() Frame {
	return frameFrom(f.nago().Grow())
}

// PreventShrink prevents, that the view shrinks in a stack.
//
//gostrict:host
func (f Frame) PreventShrink() Frame {
	return frameFrom(f.nago().PreventShrink())
}
