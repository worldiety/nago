// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package navsplitview

import (
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

type ViewID string
type NavLinks map[ViewID]core.View

func (n NavLinks) Create(id ViewID) core.View {
	return n[id]
}

type NavFnLinks map[ViewID]func(id ViewID) core.View

func (n NavFnLinks) Create(id ViewID) core.View {
	if f, ok := n[id]; ok {
		return f(id)
	}

	return nil
}

type NavFn func(id ViewID) core.View

func (f NavFn) Create(id ViewID) core.View {
	return f(id)
}

type Factory interface {
	Create(id ViewID) core.View
}

func NavigateContent(wnd core.Window, id string, view ViewID) {
	wnd.Navigation().ForwardTo(wnd.Path(), wnd.Values().Put(KindContent.queryKey(id), string(view)))
}

func NavigateDetail(wnd core.Window, id string, view ViewID) {
	wnd.Navigation().ForwardTo(wnd.Path(), wnd.Values().Put(KindDetail.queryKey(id), string(view)))
}

func NavigateSidebar(wnd core.Window, id string, view ViewID) {
	wnd.Navigation().ForwardTo(wnd.Path(), wnd.Values().Put(KindSidebar.queryKey(id), string(view)))
}

// TTwoColumn is a layout component (TwoColumn).
type TTwoColumn struct {
	nav            Factory
	frame          ui.Frame
	id             string
	defaultContent ViewID
	defaultDetail  ViewID
	bgContent      ui.Color
	bgDetail       ui.Color
	contentWidth   ui.Length
	detailWidth    ui.Length
}

func TwoColumn(nav Factory) TTwoColumn {
	return TTwoColumn{
		nav:          nav,
		contentWidth: "30rem",
		detailWidth:  "1fr",
	}
}

func (c TTwoColumn) ID(id string) TTwoColumn {
	c.id = id
	return c
}

func (c TTwoColumn) Frame(frame ui.Frame) TTwoColumn {
	c.frame = frame
	return c
}

func (c TTwoColumn) Default(content, detail ViewID) TTwoColumn {
	c.defaultContent = content
	c.defaultDetail = detail
	return c
}

func (c TTwoColumn) FullWidth() TTwoColumn {
	c.frame = c.frame.FullWidth()
	return c
}

// WidthContent sets the width of the content column.
func (c TTwoColumn) WidthContent(width ui.Length) TTwoColumn {
	c.contentWidth = width
	return c
}

func (c TTwoColumn) WidthDetail(width ui.Length) TTwoColumn {
	c.detailWidth = width
	return c
}

func (c TTwoColumn) BackgroundColorContent(bg ui.Color) TTwoColumn {
	c.bgContent = bg
	return c
}

func (c TTwoColumn) BackgroundColorDetail(bg ui.Color) TTwoColumn {
	c.bgDetail = bg
	return c
}

func (c TTwoColumn) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	detailKey := KindDetail.queryKey(c.id)
	contentId := wnd.Values()[KindContent.queryKey(c.id)]
	detailId := wnd.Values()[detailKey]

	var contentView core.View
	if contentId != "" {
		contentView = c.nav.Create(ViewID(contentId))
	} else {
		contentView = c.nav.Create(c.defaultContent)
	}

	if contentView == nil {
		contentView = alert.NotFound()
	}

	var detailView core.View
	if detailId != "" {
		detailView = c.nav.Create(ViewID(detailId))
	} else {
		detailView = c.nav.Create(c.defaultDetail)
	}

	if detailView == nil {
		detailView = alert.NotFound()
	}

	if wnd.Info().SizeClass <= core.SizeClassMedium {
		if detailView == nil {
			return ui.VStack().Render(ctx)
		}

		return ui.VStack(
			ui.If(detailId != "", ui.SecondaryButton(func() {
				wnd.Navigation().BackwardTo(wnd.Path(), wnd.Values().Delete(detailKey))
			}).PreIcon(icons.ChevronLeft).Title(rstring.ActionBack.Get(wnd))),
			ui.If(detailId == "", contentView),
			ui.If(detailId != "", detailView),
		).Alignment(ui.Leading).
			Gap(ui.L8).
			FullWidth().
			Render(ctx)
	}

	return ui.Grid(
		ui.GridCell(contentView).BackgroundColor(c.bgContent),
		ui.GridCell(ui.Text("").Underline(true)).BackgroundColor(ui.ColorIconsMuted),
		ui.GridCell(detailView).BackgroundColor(c.bgDetail),
	).Columns(3).
		Gap(ui.L8).
		Widths(c.contentWidth, "1px", c.detailWidth).
		Frame(c.frame).
		Render(ctx)
}
