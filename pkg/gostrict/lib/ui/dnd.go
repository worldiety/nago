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

// TDnDArea is an area, whose children can be dragged and dropped, see DnDArea.
type TDnDArea struct{ ref js.Ref }

func (TDnDArea) isView() {}

// DnDArea returns an area of the children, which can be dragged and dropped.
//
//gostrict:host
func DnDArea(children ...View) TDnDArea {
	return TDnDArea{ref: wrap(nui.DnDArea(viewsOf(children)...))}
}

// ID sets the unique identifier of the area.
//
//gostrict:host
func (c TDnDArea) ID(id string) TDnDArea {
	return TDnDArea{ref: wrap(get[nui.TDnDArea](c.ref).ID(id))}
}

// CanDrag lets the children be dragged.
//
//gostrict:host
func (c TDnDArea) CanDrag(canDrag bool) TDnDArea {
	return TDnDArea{ref: wrap(get[nui.TDnDArea](c.ref).CanDrag(canDrag))}
}

// CanDrop lets the area accept dropped children.
//
//gostrict:host
func (c TDnDArea) CanDrop(canDrop bool) TDnDArea {
	return TDnDArea{ref: wrap(get[nui.TDnDArea](c.ref).CanDrop(canDrop))}
}

// Droppable sets the ids, which the area accepts, when they are dropped on it.
//
//gostrict:host
func (c TDnDArea) Droppable(ids ...string) TDnDArea {
	return TDnDArea{ref: wrap(get[nui.TDnDArea](c.ref).Droppable(ids...))}
}

// Frame sets the size and the position of the area.
//
//gostrict:host
func (c TDnDArea) Frame(frame Frame) TDnDArea {
	return TDnDArea{ref: wrap(get[nui.TDnDArea](c.ref).Frame(frame.nago()))}
}

// InputValue keeps the id of the last dropped item in the state.
//
//gostrict:host
func (c TDnDArea) InputValue(state StringState) TDnDArea {
	return TDnDArea{ref: wrap(get[nui.TDnDArea](c.ref).InputValue(state.nago()))}
}
