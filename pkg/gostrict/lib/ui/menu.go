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

// TMenu is a menu, which is anchored to a view and shows groups of items, see Menu.
type TMenu struct{ ref js.Ref }

func (TMenu) isView() {}

// Menu returns a menu, which is anchored to the anchor. Empty groups are discarded.
//
//gostrict:host
func Menu(anchor View, groups ...TMenuGroup) TMenu {
	return TMenu{ref: wrap(nui.Menu(viewOf(anchor), b4GroupsOf(groups)...))}
}

// Frame sets the layout frame of the menu.
//
//gostrict:host
func (c TMenu) Frame(frame Frame) TMenu {
	return TMenu{ref: wrap(get[nui.TMenu](c.ref).Frame(frame.nago()))}
}

// Offset sets the offset of the menu from its anchor.
//
//gostrict:host
func (c TMenu) Offset(offset Length) TMenu {
	return TMenu{ref: wrap(get[nui.TMenu](c.ref).Offset(nui.Length(offset)))}
}

// TMenuGroup is a group of related menu items, see MenuGroup.
type TMenuGroup struct{ ref js.Ref }

// MenuGroup returns a group with the items.
//
//gostrict:host
func MenuGroup(items ...TMenuItem) TMenuGroup {
	return TMenuGroup{ref: wrap(nui.MenuGroup(b4ItemsOf(items)...))}
}

// CustomContent shows the content instead of the menu items.
//
//gostrict:host
func (c TMenuGroup) CustomContent(content View) TMenuGroup {
	return TMenuGroup{ref: wrap(get[nui.TMenuGroup](c.ref).CustomContent(viewOf(content)))}
}

// TMenuItem is an entry of a menu, see MenuItem.
type TMenuItem struct{ ref js.Ref }

// MenuItem returns an item, which calls action, when the user selects it, and which shows the content.
//
//gostrict:host
func MenuItem(action func(), content View) TMenuItem {
	return TMenuItem{ref: wrap(nui.MenuItem(action, viewOf(content)))}
}
