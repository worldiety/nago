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

// TMenu is a composite component (Menu).
// It displays a contextual or dropdown menu anchored to a specific view.
// The menu is organized into groups of items, where each item has content
// and an action callback. Empty groups are ignored.
type TMenu struct {
	anchor core.View    // view the menu is anchored to
	groups []TMenuGroup // groups of menu items
	frame  Frame        // layout frame for sizing and positioning
}

// Menu creates a new menu with the given anchor and groups.
// Empty groups are discarded when rendering.
func Menu(anchor core.View, groups ...TMenuGroup) TMenu {
	return TMenu{
		anchor: anchor,
		groups: groups,
	}
}

// Frame sets the layout frame of the menu.
func (c TMenu) Frame(frame Frame) TMenu {
	c.frame = frame
	return c
}

// Render builds and returns the protocol representation of the menu.
// It renders the anchor view and each non-empty group with its items.
func (c TMenu) Render(ctx core.RenderContext) core.RenderNode {
	groups := make([]proto.MenuGroup, 0, len(c.groups))
	for _, grp := range c.groups {
		if len(grp.items) == 0 {
			continue
		}

		items := make([]proto.MenuItem, 0, len(grp.items))
		for _, item := range grp.items {
			if item.content == nil {
				continue
			}
			items = append(items, proto.MenuItem{
				Action:  ctx.MountCallback(item.action),
				Content: render(ctx, item.content),
			})
		}
		groups = append(groups, proto.MenuGroup{
			Items: items,
		})
	}

	return &proto.Menu{
		Anchor: render(ctx, c.anchor),
		Groups: groups,
		Frame:  c.frame.ora(),
	}
}

// TMenuGroup is a collection of related menu items grouped together.
type TMenuGroup struct {
	items []TMenuItem // list of menu items in the group
}

// MenuGroup creates a new menu group containing the given items.
func MenuGroup(items ...TMenuItem) TMenuGroup {
	return TMenuGroup{items: items}
}

// TMenuItem represents a single menu entry with an action and content view.
type TMenuItem struct {
	action  func()    // action executed when the item is selected
	content core.View // view to render as the item content
}

// MenuItem creates a new menu item with the given action and content.
func MenuItem(action func(), content core.View) TMenuItem {
	return TMenuItem{
		action:  action,
		content: content,
	}
}
