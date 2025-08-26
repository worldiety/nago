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

// THoverGroup is a composite component (Hover Group).
// It displays one view by default and replaces it with an alternate view
// when hovered. Useful for interactive UI elements such as cards or buttons
// that reveal additional content on hover. The group supports styling with
// frame, position, border, background color, and padding.
type THoverGroup struct {
	content         core.View // default content shown when not hovered
	hoveredContent  core.View // content shown when hovered
	frame           Frame     // layout frame for sizing and positioning
	position        Position  // position of the hover group
	border          Border    // border styling
	backgroundColor Color     // background color of the container
	padding         Padding   // inner spacing around the content
}

// HoverGroup creates a new hover group with default and hovered content.
// By default, the hover group is oriented vertically.
func HoverGroup(content core.View, hoveredContent core.View) THoverGroup {
	return THoverGroup{
		content:        content,
		hoveredContent: hoveredContent,
	}
}

// Frame sets the layout frame of the hover group.
func (c THoverGroup) Frame(frame Frame) THoverGroup {
	c.frame = frame
	return c
}

// Position sets the position of the hover group.
func (c THoverGroup) Position(position Position) THoverGroup {
	c.position = position
	return c
}

// Border sets the border styling of the hover group.
func (c THoverGroup) Border(border Border) THoverGroup {
	c.border = border
	return c
}

// Padding sets the inner spacing around the hover group's content.
func (c THoverGroup) Padding(padding Padding) THoverGroup {
	c.padding = padding
	return c
}

// BackgroundColor sets the background color of the hover group.
func (c THoverGroup) BackgroundColor(color Color) THoverGroup {
	c.backgroundColor = color
	return c
}

// Render builds and returns the protocol representation of the hover group,
// including both default and hovered content, frame, position, border,
// background color, and padding.
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
