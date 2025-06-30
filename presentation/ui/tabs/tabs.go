// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tabs

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type TPage struct {
	title    string
	icon     core.SVG
	body     func() core.View
	disabled bool
}

func Page(title string, body func() core.View) TPage {
	return TPage{
		title: title,
		body:  body,
	}
}

func (c TPage) Icon(ico core.SVG) TPage {
	c.icon = ico
	return c
}

func (c TPage) Disabled(disabled bool) TPage {
	c.disabled = disabled
	return c
}

type TTabs struct {
	pages         []TPage
	frame         ui.Frame
	position      ui.Position
	tabAlignment  ui.Alignment
	idx           *core.State[int]
	pageTabSpacer ui.Length
}

func Tabs(pages ...TPage) TTabs {
	return TTabs{
		pages:         pages,
		tabAlignment:  ui.Leading,
		pageTabSpacer: ui.L32,
	}
}

func (c TTabs) Frame(frame ui.Frame) TTabs {
	c.frame = frame
	return c
}

func (c TTabs) FullWidth() TTabs {
	c.frame.Width = ui.Full
	return c
}

// ButtonAlignment sets the alignment of the tab buttons within the button bar. Defaults to Leading.
func (c TTabs) ButtonAlignment(tabAlignment ui.Alignment) TTabs {
	c.tabAlignment = tabAlignment
	return c
}

// PageTabSpace is the amount of space between the tab button bar and the actual page content. Default to L32.
// Set to the empty string to disable any space.
func (c TTabs) PageTabSpace(space ui.Length) TTabs {
	c.pageTabSpacer = space
	return c
}

func (c TTabs) Position(pos ui.Position) TTabs {
	c.position = pos
	return c
}

func (c TTabs) InputValue(activeIdx *core.State[int]) TTabs {
	c.idx = activeIdx
	return c
}

func (c TTabs) Render(ctx core.RenderContext) core.RenderNode {
	idx := -1
	return ui.VStack(
		ui.ScrollView(
			ui.HStack(
				ui.ForEach(c.pages, func(p TPage) core.View {
					idx++
					myIdx := idx
					active := c.idx != nil && c.idx.Get() == idx
					style := ui.StyleButtonSecondary
					if active {
						style = ui.StyleButtonPrimary
					}

					return ui.TertiaryButton(func() {
						if c.idx != nil {
							c.idx.Set(myIdx)
							c.idx.Notify()
						}
					}).Title(p.title).PreIcon(p.icon).Preset(style).Enabled(c.idx != nil && !p.disabled)
				})...,
			).FullWidth().Alignment(c.tabAlignment).Gap(ui.L8).Padding(ui.Padding{Bottom: ui.L8}),
		).Axis(ui.ScrollViewAxisHorizontal).Frame(ui.Frame{Width: ui.Full}),
		ui.If(c.pageTabSpacer != "", ui.Space(c.pageTabSpacer)),
		func() core.View {
			if c.idx == nil || c.idx.Get() < 0 || c.idx.Get() >= len(c.pages) {
				return nil
			}

			return c.pages[c.idx.Get()].body()
		}(),
	).Position(c.position).Frame(c.frame).Render(ctx)
}
