// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package pager

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"golang.org/x/text/language"
)

var (
	StrPreviousPage = i18n.MustString("nago.pager.previous_page", i18n.Values{language.English: "Previous page", language.German: "Vorherige Seite"})
	StrNextPage     = i18n.MustString("nago.pager.next_page", i18n.Values{language.English: "Next page", language.German: "Nächste Seite"})
	StrPageXOfY     = i18n.MustVarString("nago.pager.page_x_of_y", i18n.Values{language.English: "{page} of {count}", language.German: "{page} von {count}"})
)

// TPager is a utility component (Pager).
// It manages pagination state and renders controls to switch between pages.
// Typical usage is in lists or tables with many items.
type TPager struct {
	count     int
	page      *core.State[int]
	frame     ui.Frame
	invisible bool
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
	c.invisible = !v
	return c
}

func (c TPager) Render(ctx core.RenderContext) core.RenderNode {
	if c.invisible {
		return ui.VStack().Visible(false).Render(ctx)
	}

	wnd := ctx.Window()
	if c.count <= 1 {
		return ui.HStack(
			ui.TertiaryButton(nil).PreIcon(flowbiteOutline.ChevronLeft).AccessibilityLabel(StrPreviousPage.Get(wnd)).Enabled(false),
			ui.Text(StrPageXOfY.Get(wnd, i18n.Int("page", 1), i18n.Int("count", 1))),
			ui.TertiaryButton(nil).PreIcon(flowbiteOutline.ChevronRight).AccessibilityLabel(StrNextPage.Get(wnd)).Enabled(false),
		).Gap(ui.L16).Frame(c.frame).Render(ctx)
	}

	return ui.HStack(
		ui.TertiaryButton(func() {
			c.page.Set(c.page.Get() - 1)
			c.page.Notify()
		}).PreIcon(flowbiteOutline.ChevronLeft).AccessibilityLabel(StrPreviousPage.Get(wnd)).Enabled(c.page.Get() > 0),
		ui.Text(StrPageXOfY.Get(wnd, i18n.Int("page", c.page.Get()+1), i18n.Int("count", c.count))),
		ui.TertiaryButton(func() {
			c.page.Set(c.page.Get() + 1)
			c.page.Notify()
		}).PreIcon(flowbiteOutline.ChevronRight).AccessibilityLabel(StrNextPage.Get(wnd)).Enabled(c.page.Get() < c.count-1),
	).Gap(ui.L8).Frame(c.frame).Render(ctx)
}
