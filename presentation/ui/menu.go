package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TMenu struct {
	anchor core.View
	groups []TMenuGroup
}

func Menu(anchor core.View, groups ...TMenuGroup) TMenu {
	return TMenu{
		anchor: anchor,
		groups: groups,
	}
}

func (c TMenu) Render(ctx core.RenderContext) core.RenderNode {
	groups := make([]proto.MenuGroup, 0, len(c.groups))
	for _, grp := range c.groups {
		items := make([]proto.MenuItem, 0, len(grp.Items))
		for _, item := range grp.Items {
			items = append(items, proto.MenuItem{
				Action:  ctx.MountCallback(item.Action),
				Content: render(ctx, item.Content),
			})
		}
		groups = append(groups, proto.MenuGroup{
			Items: items,
		})
	}

	return &proto.Menu{
		Anchor: render(ctx, c.anchor),
		Groups: groups,
	}
}

type TMenuGroup struct {
	Items []TMenuItem
}

func MenuGroup(items ...TMenuItem) TMenuGroup {
	return TMenuGroup{Items: items}
}

type TMenuItem struct {
	Action  func()
	Content core.View
}

func MenuItem(action func(), content core.View) TMenuItem {
	return TMenuItem{
		Action:  action,
		Content: content,
	}
}
