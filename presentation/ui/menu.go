package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TMenu struct {
	anchor core.View
	groups []TMenuGroup
	frame  Frame
}

func Menu(anchor core.View, groups ...TMenuGroup) TMenu {
	return TMenu{
		anchor: anchor,
		groups: groups,
	}
}

func (c TMenu) Frame(frame Frame) TMenu {
	c.frame = frame
	return c
}

func (c TMenu) Render(ctx core.RenderContext) core.RenderNode {
	groups := make([]proto.MenuGroup, 0, len(c.groups))
	for _, grp := range c.groups {
		items := make([]proto.MenuItem, 0, len(grp.items))
		for _, item := range grp.items {
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

type TMenuGroup struct {
	items []TMenuItem
}

func MenuGroup(items ...TMenuItem) TMenuGroup {
	return TMenuGroup{items: items}
}

type TMenuItem struct {
	action  func()
	content core.View
}

func MenuItem(action func(), content core.View) TMenuItem {
	return TMenuItem{
		action:  action,
		content: content,
	}
}
