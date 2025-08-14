// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package pager

import (
	"fmt"

	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

type TPager struct {
	count   int
	page    *core.State[int]
	frame   ui.Frame
	visible bool
}

// Pager creates a new Page with the given state.
// The active page is the zero-based offset of the current page.
func Pager(pageIdx *core.State[int]) TPager {
	return TPager{
		page: pageIdx,
	}
}

// Count sets the number of available pages. If count is 0 or less, at least a single page is still shown.
func (c TPager) Count(count int) TPager {
	c.count = count
	return c
}

func (c TPager) Frame(frame ui.Frame) TPager {
	c.frame = frame
	return c
}

func (c TPager) Visible(v bool) TPager {
	c.visible = v
	return c
}

func (c TPager) Render(ctx core.RenderContext) core.RenderNode {
	if !c.visible {
		return ui.VStack().Visible(false).Render(ctx)
	}

	if c.count <= 1 {
		return ui.HStack(
			ui.TertiaryButton(nil).PreIcon(flowbiteOutline.ChevronLeft).Enabled(false),
			ui.Text("1 von 1"),
			ui.TertiaryButton(nil).PreIcon(flowbiteOutline.ChevronRight).Enabled(false),
		).Gap(ui.L16).Frame(c.frame).Render(ctx)
	}

	return ui.HStack(
		ui.TertiaryButton(func() {
			c.page.Set(c.page.Get() - 1)
			c.page.Notify()
		}).PreIcon(flowbiteOutline.ChevronLeft).Enabled(c.page.Get() > 0),
		ui.Text(fmt.Sprintf("%d von %d", c.page.Get()+1, c.count)),
		ui.TertiaryButton(func() {
			c.page.Set(c.page.Get() + 1)
			c.page.Notify()
		}).PreIcon(flowbiteOutline.ChevronRight).Enabled(c.page.Get() < c.count-1),
	).Gap(ui.L8).Frame(c.frame).Render(ctx)
}
