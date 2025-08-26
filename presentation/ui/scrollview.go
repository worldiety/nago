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

type ScrollViewAxis int

func (a ScrollViewAxis) ora() proto.ScrollViewAxis {
	return proto.ScrollViewAxis(a)
}

const (
	ScrollViewAxisVertical   ScrollViewAxis = ScrollViewAxis(proto.ScrollViewAxisVertical)
	ScrollViewAxisHorizontal                = ScrollViewAxis(proto.ScrollViewAxisHorizontal)
	ScrollViewAxisBoth                      = ScrollViewAxis(proto.ScrollViewAxisBoth)
)

// TScrollView is a composite component (Scroll View).
// It provides a scrollable container for a single child view.
// The scroll direction can be vertical (default) or horizontal.
// Supports customization of frame, position, border, background color, and padding.
type TScrollView struct {
	content         core.View      // the scrollable content
	axis            ScrollViewAxis // scroll direction (vertical/horizontal)
	frame           Frame          // layout frame for size and positioning
	position        Position       // content alignment within the scroll view
	border          Border         // optional border around the scroll view
	backgroundColor Color          // background color
	padding         Padding        // inner padding
}

// A ScrollView can either be horizontal or vertical. By default, it is vertical.
func ScrollView(content core.View) TScrollView {
	return TScrollView{
		content: content,
		axis:    ScrollViewAxisVertical,
	}
}

// Axis sets the scroll direction (vertical or horizontal).
func (c TScrollView) Axis(axis ScrollViewAxis) TScrollView {
	c.axis = axis
	return c
}

// Frame sets the layout frame for the scroll view.
func (c TScrollView) Frame(frame Frame) TScrollView {
	c.frame = frame
	return c
}

// Position sets the alignment of the content inside the scroll view.
func (c TScrollView) Position(position Position) TScrollView {
	c.position = position
	return c
}

// Border applies a border around the scroll view.
func (c TScrollView) Border(border Border) TScrollView {
	c.border = border
	return c
}

// Padding sets the inner padding of the scroll view.
func (c TScrollView) Padding(padding Padding) TScrollView {
	c.padding = padding
	return c
}

// BackgroundColor sets the background color of the scroll view.
func (c TScrollView) BackgroundColor(color Color) TScrollView {
	c.backgroundColor = color
	return c
}

// Render builds and returns the protocol representation of the scroll view.
// It includes the scrollable content, axis (vertical/horizontal),
// frame, alignment, border, background color, and padding.
func (c TScrollView) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.ScrollView{
		Content:         render(ctx, c.content),
		Axis:            c.axis.ora(),
		Frame:           c.frame.ora(),
		Position:        c.position.ora(),
		Border:          c.border.ora(),
		BackgroundColor: c.backgroundColor.ora(),
		Padding:         c.padding.ora(),
	}
}
