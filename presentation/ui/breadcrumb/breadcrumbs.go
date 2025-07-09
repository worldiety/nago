// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package breadcrumb

import (
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

type TBreadcrumbs struct {
	items []core.View
	gap   ui.Length
	frame ui.Frame
}

func Breadcrumbs(items ...core.View) TBreadcrumbs {
	return TBreadcrumbs{items: items}
}

func (c TBreadcrumbs) Gap(l ui.Length) TBreadcrumbs {
	c.gap = l
	return c
}

func (c TBreadcrumbs) Frame(frame ui.Frame) TBreadcrumbs {
	c.frame = frame
	return c
}

// Item appends a default button with the given title and text. Currently, this just defaults
// to a tertiary styled button.
func (c TBreadcrumbs) Item(title string, action func()) TBreadcrumbs {
	c.items = append(c.items, ui.TertiaryButton(action).Title(title))
	return c
}

func (c TBreadcrumbs) Render(ctx core.RenderContext) core.RenderNode {
	var tmp []core.View
	for idx, item := range c.items {
		tmp = append(tmp, item)
		if idx < len(c.items)-1 {
			tmp = append(tmp, ui.ImageIcon(icons.ChevronRight))
		}
	}

	return ui.HStack(tmp...).Gap(c.gap).Frame(c.frame).Render(ctx)
}
