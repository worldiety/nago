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

type TDnDArea struct {
	frame     Frame
	canDrag   bool
	canDrop   bool
	droppable []string
	children  []core.View
	id        string
	dropped   *core.State[string]
}

func DnDArea(children ...core.View) TDnDArea {
	return TDnDArea{
		children: children,
	}
}

func (c TDnDArea) ID(id string) TDnDArea {
	c.id = id
	return c
}

func (c TDnDArea) CanDrag(canDrag bool) TDnDArea {
	c.canDrag = canDrag
	return c
}

func (c TDnDArea) CanDrop(canDrop bool) TDnDArea {
	c.canDrop = canDrop
	return c
}

func (c TDnDArea) Droppable(ids ...string) TDnDArea {
	c.droppable = ids
	return c
}

func (c TDnDArea) Frame(frame Frame) TDnDArea {
	c.frame = frame
	return c
}

func (c TDnDArea) InputValue(state *core.State[string]) TDnDArea {
	c.dropped = state
	return c
}

func (c TDnDArea) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.DnDArea{
		DnD: proto.DnD{
			CanDrop:      proto.Bool(c.canDrop),
			CanDrag:      proto.Bool(c.canDrag),
			DroppableIDs: proto.NewStrings(c.droppable),
		},
		Id:        proto.Str(c.id),
		Frame:     c.frame.ora(),
		Children:  renderComponents(ctx, c.children),
		DroppedId: c.dropped.Ptr(),
	}
}
