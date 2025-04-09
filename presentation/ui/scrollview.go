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

type TScrollView struct {
	content         core.View
	axis            ScrollViewAxis
	frame           Frame
	position        Position
	border          Border
	backgroundColor Color
	padding         Padding
}

// A ScrollView can either be horizontal or vertical. By default, it is vertical.
func ScrollView(content core.View) TScrollView {
	return TScrollView{
		content: content,
		axis:    ScrollViewAxisVertical,
	}
}

func (c TScrollView) Axis(axis ScrollViewAxis) TScrollView {
	c.axis = axis
	return c
}

func (c TScrollView) Frame(frame Frame) TScrollView {
	c.frame = frame
	return c
}

func (c TScrollView) Position(position Position) TScrollView {
	c.position = position
	return c
}

func (c TScrollView) Border(border Border) TScrollView {
	c.border = border
	return c
}

func (c TScrollView) Padding(padding Padding) TScrollView {
	c.padding = padding
	return c
}

func (c TScrollView) BackgroundColor(color Color) TScrollView {
	c.backgroundColor = color
	return c
}

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
