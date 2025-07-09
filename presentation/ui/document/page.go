// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type Size struct {
	Width  ui.Length
	Height ui.Length
}

var (
	DinA4 = Size{
		Width:  "50rem", // 21cm
		Height: "70rem", // 29,7cm
	}
)

type TPage struct {
	frame           ui.Frame
	padding         ui.Padding
	border          ui.Border
	backgroundColor ui.Color
	alignment       ui.Alignment
	items           []core.View
	comments        []TComment
}

// Page create a kind of virtual endless DinA4 styled paper document.
func Page(items ...core.View) TPage {
	return TPage{
		alignment: ui.Leading,
		frame:     ui.Frame{MaxWidth: "70rem"}.FullWidth(),
		padding:   ui.Padding{}.All(ui.L48),
		items:     items,
	}
}

func (c TPage) Size(size Size) TPage {
	c.frame = ui.Frame{
		MaxWidth:  size.Width,
		MaxHeight: size.Height,
		MinWidth:  size.Width,
		MinHeight: size.Height,
		Width:     size.Width,
		Height:    size.Height,
	}

	return c
}

// Comment appends the given comments to any already existing comments.
func (c TPage) Comment(comments ...TComment) TPage {
	c.comments = append(c.comments, comments...)
	return c
}

func (c TPage) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	if wnd.Info().SizeClass <= core.SizeClassSmall {
		return ui.VStack(c.items...).
			BackgroundColor(c.backgroundColor).
			Frame(ui.Frame{Width: ui.Full}). // ignore frame by intention, that will just break
			Render(ctx)
	}

	canShowComments := wnd.Info().SizeClass > core.SizeClassMedium

	bgColor := c.backgroundColor
	if bgColor == "" {
		if wnd.Info().ColorScheme == core.Dark {
			bgColor = ui.M1
		} else {
			bgColor = ui.ColorWhite
		}
	}

	return ui.HStack(
		// actual page
		ui.VStack(
			c.items...,
		).FullWidth().
			BackgroundColor(bgColor).
			Frame(c.frame).
			Padding(ui.Padding{}.Horizontal("5rem").Vertical("4rem")).
			Border(ui.Border{}.Width(ui.L1).Color(ui.ColorIconsMuted).Radius(ui.L4)),

		// comments
		ui.IfFunc(len(c.comments) > 0 && canShowComments, func() core.View {
			var tmp []core.View
			var allMySelections []*core.State[bool]

			for idx, comment := range c.comments {
				if idx > 0 && comment.needsSepLine {
					tmp = append(tmp, ui.HLine())
				} else {
					tmp = append(tmp, ui.Space(ui.L16))
				}

				tmp = append(tmp, comment)
				if comment.selected != nil {
					allMySelections = append(allMySelections, comment.selected)
					comment.selected.Observe(func(newValue bool) {
						for _, selection := range allMySelections {
							if selection == comment.selected {
								selection.Set(newValue)
							} else {
								selection.Set(false)
							}
						}
					})
				}
			}

			// TODO what about paging?

			return ui.VStack(tmp...).
				NoClip(true).
				Alignment(ui.TopLeading).
				Frame(ui.Frame{Width: ui.L320, MaxWidth: ui.L320})
		}),
	).FullWidth().Alignment(ui.Top).Gap(ui.L32).Render(ctx)
}
