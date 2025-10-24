// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import "go.wdy.de/nago/presentation/proto"

// Frame is a layout component(Frame).
// Frame defines the sizing constraints and fixed dimensions for a UI element.
// It allows you to specify minimum and maximum width/height, as well as fixed
// dimensions. Frames are used to control layout behavior and responsiveness.
// All fields are optional. If a field is zero, it will not constrain the layout.
type Frame struct {
	MinWidth  Length
	MaxWidth  Length
	MinHeight Length
	MaxHeight Length
	Width     Length
	Height    Length
}

// IsZero returns true if all fields of the Frame are unset (zero value).
func (f Frame) IsZero() bool {
	return Frame{} == f
}

func (f Frame) ora() proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(f.MinWidth),
		MaxWidth:  proto.Length(f.MaxWidth),
		MinHeight: proto.Length(f.MinHeight),
		MaxHeight: proto.Length(f.MaxHeight),
		Width:     proto.Length(f.Width),
		Height:    proto.Length(f.Height),
	}
}

// Size sets both Width and Height to the given values and returns the updated Frame.
func (f Frame) Size(w, h Length) Frame {
	f.Height = h
	f.Width = w
	return f
}

// MatchScreen sets the frame to match the full viewport height and width.
// This is useful for fullscreen layouts or sections that should fill the screen.
func (f Frame) MatchScreen() Frame {
	f.MinHeight = ViewportHeight
	f.Width = Full
	return f
}

// FullWidth sets the frame's width to 100% of the available space.
func (f Frame) FullWidth() Frame {
	f.Width = "100%"
	return f
}

// Large sets the max width to 560dp (35rem) and Width to Full.
func (f Frame) Large() Frame {
	f.MaxWidth = L560
	f.Width = Full
	return f
}

// Larger sets the width to 880dp (55rem) and Width to Full.
func (f Frame) Larger() Frame {
	f.MaxWidth = L880
	f.Width = Full
	return f
}

// FullHeight sets the frame's height to 100% of the available space.
func (f Frame) FullHeight() Frame {
	f.Height = "100%"
	return f
}
