// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package navsplitview

import (
	"fmt"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type TargetKind int

func (k TargetKind) queryKey(id string) string {
	switch k {
	case KindSidebar:
		return id + "_nav_sidebar"
	case KindContent:
		return id + "_nav_content"

	case KindDetail:
		return id + "_nav_detail"

	default:
		panic(fmt.Sprintf("unknown target kind %d", k))
	}

}

const (
	KindSidebar TargetKind = iota + 1
	KindContent
	KindDetail
)

type TListItem struct {
	content  core.View
	kind     TargetKind
	frame    ui.Frame
	target   ViewID
	idPrefix string
}

func ListItem(kind TargetKind, target ViewID, content core.View) TListItem {
	return TListItem{
		content: content,
		kind:    kind,
		frame:   ui.Frame{}.FullWidth(),
		target:  target,
	}
}

func (c TListItem) Frame(frame ui.Frame) TListItem {
	c.frame = frame
	return c
}

func (c TListItem) Prefix(id string) TListItem {
	c.idPrefix = id
	return c
}

func (c TListItem) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	key := c.kind.queryKey(c.idPrefix)
	selected := string(c.target) == wnd.Values()[key]

	var bgColor ui.Color
	if selected {
		bgColor = ui.ColorCardFooter
	}

	return ui.HStack(c.content).
		BackgroundColor(bgColor).
		HoveredBackgroundColor(ui.ColorCardFooter).
		Action(func() {
			wnd.Navigation().ForwardTo(wnd.Path(), wnd.Values().Put(key, string(c.target)))
		}).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8)).
		Frame(c.frame).
		Render(ctx)
}
