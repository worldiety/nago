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

// TBreadcrumbs is a feedback component (Breadcrumbs).
// It displays a horizontal trail of items representing the user's navigation path
// within the application. Each item is typically a link or label, separated by a
// configurable gap, and the layout can be styled with frame and padding options.
type TBreadcrumbs struct {
	items   []core.View
	gap     ui.Length
	frame   ui.Frame
	padding ui.Padding
}

// Breadcrumbs creates a new breadcrumb trail with the given items.
func Breadcrumbs(items ...core.View) TBreadcrumbs {
	return TBreadcrumbs{items: items}
}

// Gap sets the spacing between breadcrumb items.
func (c TBreadcrumbs) Gap(l ui.Length) TBreadcrumbs {
	c.gap = l
	return c
}

// Frame defines the frame layout (size and positioning) of the breadcrumbs.
func (c TBreadcrumbs) Frame(frame ui.Frame) TBreadcrumbs {
	c.frame = frame
	return c
}

// Padding sets the inner padding around the breadcrumb trail.
func (c TBreadcrumbs) Padding(padding ui.Padding) TBreadcrumbs {
	c.padding = padding
	return c
}

// ClampLeading ensures that if the first entry is a default Item its title will be aligned to the
// leading of this component so that you can align the optical flight of text.
func (c TBreadcrumbs) ClampLeading() TBreadcrumbs {
	c.padding.Left = "-1.2rem"
	return c
}

// Item appends a default button with the given title and text. Currently, this just defaults
// to a tertiary styled button.
func (c TBreadcrumbs) Item(title string, action func()) TBreadcrumbs {
	c.items = append(c.items, ui.TertiaryButton(action).Title(title))
	return c
}

// Render builds the breadcrumb trail as a horizontal stack,
// automatically inserting a chevron (›) icon between items
// and applying the configured gap, frame, and padding.
func (c TBreadcrumbs) Render(ctx core.RenderContext) core.RenderNode {
	var tmp []core.View
	for idx, item := range c.items {
		tmp = append(tmp, item)
		if idx < len(c.items)-1 {
			tmp = append(tmp, ui.ImageIcon(icons.ChevronRight))
		}
	}

	return ui.HStack(tmp...).Gap(c.gap).Frame(c.frame).Padding(c.padding).Render(ctx)
}
