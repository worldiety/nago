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

type THoverGroup struct {
	content         core.View
	hoveredContent  core.View
	frame           Frame
	position        Position
	border          Border
	backgroundColor Color
	padding         Padding
}

// A HoverGroup can either be horizontal or vertical. By default, it is vertical.
func HoverGroup(content core.View, hoveredContent core.View) THoverGroup {
	return THoverGroup{
		content:        content,
		hoveredContent: hoveredContent,
	}
}

func (c THoverGroup) Frame(frame Frame) THoverGroup {
	c.frame = frame
	return c
}

func (c THoverGroup) Position(position Position) THoverGroup {
	c.position = position
	return c
}

func (c THoverGroup) Border(border Border) THoverGroup {
	c.border = border
	return c
}

func (c THoverGroup) Padding(padding Padding) THoverGroup {
	c.padding = padding
	return c
}

func (c THoverGroup) BackgroundColor(color Color) THoverGroup {
	c.backgroundColor = color
	return c
}

func (c THoverGroup) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.HoverGroup{
		Content:         render(ctx, c.content),
		ContentHover:    render(ctx, c.hoveredContent),
		Frame:           c.frame.ora(),
		Position:        c.position.ora(),
		Border:          c.border.ora(),
		BackgroundColor: c.backgroundColor.ora(),
		Padding:         c.padding.ora(),
	}
}
