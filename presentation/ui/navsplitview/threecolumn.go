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

// TThreeColumn is a layout component (TwoColumn).
type TThreeColumn struct {
	nav              Factory
	frame            ui.Frame
	id               string
	defaultContent   ViewID
	defaultDetail    ViewID
	defaultSidebar   ViewID
	bgContent        ui.Color
	bgDetail         ui.Color
	bgSidebar        ui.Color
	contentWidth     ui.Length
	sidebarWidth     ui.Length
	detailWidth      ui.Length
	alignmentContent ui.Alignment
	alignmentSidebar ui.Alignment
	alignmentDetail  ui.Alignment
}

// ThreeColumn creates a three-column layout with the given navigation factory. The order is from left to
// right: sidebar, content, detail.
func ThreeColumn(nav Factory) TThreeColumn {
	return TThreeColumn{
		nav:          nav,
		contentWidth: "25rem",
		sidebarWidth: "20rem",
		detailWidth:  "1fr",
	}
}

func (c TThreeColumn) ID(id string) TThreeColumn {
	c.id = id
	return c
}

func (c TThreeColumn) Frame(frame ui.Frame) TThreeColumn {
	c.frame = frame
	return c
}

func (c TThreeColumn) Default(sidebar, content, detail ViewID) TThreeColumn {
	c.defaultContent = content
	c.defaultDetail = detail
	c.defaultSidebar = sidebar
	return c
}

func (c TThreeColumn) FullWidth() TThreeColumn {
	c.frame = c.frame.FullWidth()
	return c
}

// WidthContent sets the width of the content column.
func (c TThreeColumn) WidthContent(width ui.Length) TThreeColumn {
	c.contentWidth = width
	return c
}

func (c TThreeColumn) WidthSidebar(width ui.Length) TThreeColumn {
	c.sidebarWidth = width
	return c
}
func (c TThreeColumn) WidthDetail(width ui.Length) TThreeColumn {
	c.sidebarWidth = width
	return c
}

func (c TThreeColumn) BackgroundColorContent(bg ui.Color) TThreeColumn {
	c.bgContent = bg
	return c
}

func (c TThreeColumn) BackgroundColorDetail(bg ui.Color) TThreeColumn {
	c.bgDetail = bg
	return c
}

func (c TThreeColumn) BackgroundColorSidebar(bg ui.Color) TThreeColumn {
	c.bgSidebar = bg
	return c
}

func (c TThreeColumn) AlignmentContent(alignment ui.Alignment) TThreeColumn {
	c.alignmentContent = alignment
	return c
}

func (c TThreeColumn) AlignmentDetail(alignment ui.Alignment) TThreeColumn {
	c.alignmentDetail = alignment
	return c
}

func (c TThreeColumn) AlignmentSidebar(alignment ui.Alignment) TThreeColumn {
	c.alignmentSidebar = alignment
	return c
}

func (c TThreeColumn) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	detailKey := KindDetail.queryKey(c.id)
	contentKey := KindContent.queryKey(c.id)
	contentId := wnd.Values()[KindContent.queryKey(c.id)]
	detailId := wnd.Values()[detailKey]
	sidebarID := wnd.Values()[KindSidebar.queryKey(c.id)]

	var sidebarView core.View
	if sidebarID != "" {
		sidebarView = c.nav.Create(ViewID(sidebarID))
	} else {
		sidebarView = c.nav.Create(c.defaultSidebar)
	}

	if sidebarView == nil {
		sidebarView = alert.NotFound()
	}

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

	// smartphone etc
	if wnd.Info().SizeClass <= core.SizeClassMedium {
		if detailView == nil {
			return ui.VStack().Render(ctx)
		}

		var myView core.View
		var backAction func()
		switch {
		case detailId != "":
			myView = detailView
			backAction = func() {
				wnd.Navigation().BackwardTo(wnd.Path(), wnd.Values().Delete(detailKey))
			}
		case contentId != "":
			myView = contentView
			backAction = func() {
				wnd.Navigation().BackwardTo(wnd.Path(), wnd.Values().Delete(contentKey))
			}
		default:
			myView = sidebarView
		}

		return ui.VStack(
			ui.If(backAction != nil, ui.SecondaryButton(backAction).
				PreIcon(icons.ChevronLeft).Title(rstring.ActionBack.Get(wnd))),

			myView,
		).Alignment(ui.Leading).
			Gap(ui.L8).
			FullWidth().
			Render(ctx)
	}

	// collapse sidebar on small screens
	if wnd.Info().SizeClass <= core.SizeClassLarge {
		collapsedLeft := core.StateOf[bool](wnd, "collapsedLeft-"+c.id).Init(func() bool {
			return true
		})
		collapsedSize := ui.L48
		detailWidth := c.detailWidth
		contentWidth := c.contentWidth
		if !collapsedLeft.Get() {
			collapsedSize = c.sidebarWidth

			detailWidth = ui.L48
			contentWidth = "1fr"
		}

		collapsedLeftIcon := icons.ChevronLeft
		if collapsedLeft.Get() {
			collapsedLeftIcon = icons.ChevronRight

		}

		return ui.Grid(
			ui.GridCell(
				ui.VStack(
					ui.VStack(
						ui.TertiaryButton(func() {
							collapsedLeft.Set(!collapsedLeft.Get())
						}).PreIcon(collapsedLeftIcon),
					).FullWidth().Alignment(ui.Trailing),
					ui.If(!collapsedLeft.Get(),
						sidebarView,
					),
				).Alignment(ui.Top).FullWidth(),
			).BackgroundColor(c.bgSidebar),
			ui.GridCell(ui.Text("").Underline(true)).BackgroundColor(ui.ColorIconsMuted),
			ui.GridCell(contentView).BackgroundColor(c.bgContent),
			ui.GridCell(ui.Text("").Underline(true)).BackgroundColor(ui.ColorIconsMuted),
			ui.GridCell(ui.VStack(
				ui.If(collapsedLeft.Get(), detailView),
				ui.If(!collapsedLeft.Get(), ui.TertiaryButton(func() {
					collapsedLeft.Set(!collapsedLeft.Get())
				}).PreIcon(icons.ChevronLeft)),
			)).BackgroundColor(c.bgDetail),
		).Columns(5).
			Gap(ui.L8).
			Widths(collapsedSize, "1px", contentWidth, "1px", detailWidth).
			Frame(c.frame).
			Render(ctx)
	}

	// enough space for all
	return ui.Grid(
		ui.GridCell(sidebarView).BackgroundColor(c.bgSidebar).Alignment(c.alignmentSidebar),
		ui.GridCell(ui.Text("").Underline(true)).BackgroundColor(ui.ColorIconsMuted),
		ui.GridCell(contentView).BackgroundColor(c.bgContent).Alignment(c.alignmentContent),
		ui.GridCell(ui.Text("").Underline(true)).BackgroundColor(ui.ColorIconsMuted),
		ui.GridCell(detailView).BackgroundColor(c.bgDetail).Alignment(c.alignmentDetail),
	).Columns(5).
		Gap(ui.L8).
		Widths(c.sidebarWidth, "1px", c.contentWidth, "1px", c.detailWidth).
		Frame(c.frame).
		Render(ctx)
}
