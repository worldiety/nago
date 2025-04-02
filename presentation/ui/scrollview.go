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
	content core.View
	axis    ScrollViewAxis
	frame   Frame
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

func (c TScrollView) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.ScrollView{
		Content: render(ctx, c.content),
		Axis:    c.axis.ora(),
		Frame:   c.frame.ora(),
	}
}
