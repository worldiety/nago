// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)


// TSpace is a layout component (Space).
// It represents a fixed-size spacer used to add consistent spacing
// between UI elements in both vertical and horizontal layouts.
type TSpace struct {
	size Length // the fixed dimension applied to both width and height
}

// Space creates a fixed-size spacer with the given length.
func Space(size Length) TSpace {
	return TSpace{size: size}
}

// Render builds the protocol representation of the space component,
// enforcing both width and height to the given size.
func (t TSpace) Render(ctx core.RenderContext) core.RenderNode {
	return VStack().Frame(Frame{MinWidth: t.size, MaxWidth: t.size, MinHeight: t.size, MaxHeight: t.size}).Render(ctx)
}


// TSpacer is a layout component (Spacer).
// Unlike TSpace, Spacer grows and shrinks dynamically to fill
// available space inside VStack or HStack containers.
type TSpacer struct {
	backgroundColor proto.Color  // optional background color
	frame           proto.Frame  // optional frame constraints
	border          proto.Border // optional border styling
}

// Spacer creates a dynamic spacer that expands to fill available space.
func Spacer() TSpacer {
	return TSpacer{}
}

// BackgroundColor sets the background color of the spacer.
func (c TSpacer) BackgroundColor(backgroundColor Color) TSpacer {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// Frame sets the frame of the spacer, allowing control over its layout constraints.
func (c TSpacer) Frame(frame Frame) TSpacer {
	c.frame = frame.ora()
	return c
}

// Border sets the border of the spacer.
func (c TSpacer) Border(border Border) {
	c.border = border.ora()
}

// Render builds the protocol representation of the spacer,
// including its frame, background color, and optional border.
func (c TSpacer) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.Spacer{
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Border:          c.border,
	}
}
