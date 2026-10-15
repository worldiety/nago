// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// THoverGroup is an overlay component, see HoverGroup. It displays one view by default and replaces it with an
// alternate view, while the pointer is over it.
type THoverGroup struct{ ref js.Ref }

func (THoverGroup) isView() {}

// HoverGroup returns a group, which shows content by default and hoveredContent, while the pointer is over it, e.g.
// to reveal additional content on hover.
//
//gostrict:host
func HoverGroup(content View, hoveredContent View) THoverGroup {
	return THoverGroup{ref: wrap(nui.HoverGroup(viewOf(content), viewOf(hoveredContent)))}
}

// Frame sets the layout frame of the hover group.
//
//gostrict:host
func (c THoverGroup) Frame(frame Frame) THoverGroup {
	return THoverGroup{ref: wrap(get[nui.THoverGroup](c.ref).Frame(frame.nago()))}
}

// Position sets the position of the hover group.
//
//gostrict:host
func (c THoverGroup) Position(position Position) THoverGroup {
	return THoverGroup{ref: wrap(get[nui.THoverGroup](c.ref).Position(position.nago()))}
}

// Border sets the border styling of the hover group.
//
//gostrict:host
func (c THoverGroup) Border(border Border) THoverGroup {
	return THoverGroup{ref: wrap(get[nui.THoverGroup](c.ref).Border(border.nago()))}
}

// Padding sets the inner spacing around the content of the hover group.
//
//gostrict:host
func (c THoverGroup) Padding(padding Padding) THoverGroup {
	return THoverGroup{ref: wrap(get[nui.THoverGroup](c.ref).Padding(padding.nago()))}
}

// BackgroundColor sets the background color of the hover group.
//
//gostrict:host
func (c THoverGroup) BackgroundColor(color Color) THoverGroup {
	return THoverGroup{ref: wrap(get[nui.THoverGroup](c.ref).BackgroundColor(nui.Color(color)))}
}
