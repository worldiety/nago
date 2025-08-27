// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tabs

import (
	"strconv"
	"unicode/utf8"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// TPage is a utility component (Page).
// Each page can define a title, an optional icon, and a body
// function that renders the content. Pages can also be disabled
// to prevent user interaction.
type TPage struct {
	title    string           // text label shown on the tab
	icon     core.SVG         // optional icon displayed next to the title
	body     func() core.View // function that renders the page content
	disabled bool             // when true, the tab is visible but not selectable
}

// Page creates a new tab page with the given title and content body.
func Page(title string, body func() core.View) TPage {
	return TPage{
		title: title,
		body:  body,
	}
}

// Icon sets the icon displayed next to the page title.
func (c TPage) Icon(ico core.SVG) TPage {
	c.icon = ico
	return c
}

// Disabled marks the page as disabled, making it visible but not selectable.
func (c TPage) Disabled(disabled bool) TPage {
	c.disabled = disabled
	return c
}

// TTabs is a utility component (Tabs).
// It manages the layout and navigation between different TPage elements,
// including alignment, positioning, and spacing between the tab bar and content.
// An optional state can track the currently active tab index.
type TTabs struct {
	pages         []TPage          // list of tab pages
	frame         ui.Frame         // layout frame for the entire tabs container
	position      ui.Position      // position of the tab bar (e.g., top, bottom)
	tabAlignment  ui.Alignment     // alignment of the tab buttons (e.g., leading, center)
	idx           *core.State[int] // external or internal state tracking the active page index
	pageTabSpacer ui.Length        // spacing between the tab bar and the page content
}

// Tabs creates a new tab container with the given pages,
// defaulting to leading alignment and a standard page-to-tab spacer.
func Tabs(pages ...TPage) TTabs {
	return TTabs{
		pages:         pages,
		tabAlignment:  ui.Leading,
		pageTabSpacer: ui.L32,
	}
}

// Frame sets the layout frame of the tabs container, including size and spacing.
func (c TTabs) Frame(frame ui.Frame) TTabs {
	c.frame = frame
	return c
}

// FullWidth sets the tabs container to span the full available width.
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

// Position sets the position of the tab bar (e.g., top, bottom, start, end).
func (c TTabs) Position(pos ui.Position) TTabs {
	c.position = pos
	return c
}

// InputValue binds the tab container to an external state that tracks
// the index of the currently active page.
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
							if utf8.ValidString(c.idx.ID()) {
								ctx.Window().Navigation().ForwardTo(ctx.Window().Path(), ctx.Window().Values().Put(c.idx.ID()+"-idx", strconv.Itoa(myIdx)))
							}
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

// NewIndexState uses [core.StateOf] to create a new state but it is initialized using the name and the postfix
// -idx to pass the index through the query parameter. If a valid name is used, clicking the page tab button will
// cause a navigation.
func NewIndexState(wnd core.Window, name string) *core.State[int] {
	return core.StateOf[int](wnd, name).Init(func() int {
		idx, _ := strconv.Atoi(wnd.Values()[name+"-idx"])
		return idx
	})
}
