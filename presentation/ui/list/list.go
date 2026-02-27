// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package list

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// TEntry is a composite component (Entry).
// Represents a single row or list item with optional headline,
// supporting text/view, leading & trailing views, and an action handler.
type TEntry struct {
	headline       string // main title text
	headlineView   core.View
	supportingText string    // optional supporting text
	supportingView core.View // optional supporting view
	leading        core.View // optional leading icon/view
	trailing       core.View // optional trailing icon/view
	action         func()    // optional action on click/tap
	frame          ui.Frame  // layout frame
}

// Entry creates a new full-width entry with default frame.
func Entry() TEntry {
	return TEntry{}.Frame(ui.Frame{}.FullWidth())
}

// Headline sets the main title text of the entry.
func (c TEntry) Headline(s string) TEntry {
	c.headline = s
	return c
}

// SupportingText sets an optional supporting text below the headline.
func (c TEntry) SupportingText(s string) TEntry {
	c.supportingText = s
	return c
}

// SupportingView sets an optional supporting view below the headline.
func (c TEntry) SupportingView(view core.View) TEntry {
	c.supportingView = view
	return c
}

func (c TEntry) HeadlineView(view core.View) TEntry {
	c.headlineView = view
	return c
}

// Leading sets an optional leading view (e.g. icon/avatar).
func (c TEntry) Leading(v core.View) TEntry {
	c.leading = v
	return c
}

// Trailing sets an optional trailing view (e.g. button/chevron).
func (c TEntry) Trailing(v core.View) TEntry {
	c.trailing = v
	return c
}

// Action sets a click/tap action handler.
func (c TEntry) Action(fn func()) TEntry {
	c.action = fn
	return c
}

// Frame sets the layout frame for the entry.
func (c TEntry) Frame(frame ui.Frame) TEntry {
	c.frame = frame
	return c
}

// Render builds the entry layout with optional leading, headline,
// supporting text/view, trailing view and click action.
func (c TEntry) Render(ctx core.RenderContext) core.RenderNode {

	return ui.HStack(
		c.leading,
		ui.VStack(
			ui.If(c.headline != "", ui.Text(c.headline).Font(ui.SubTitle)),
			ui.If(c.supportingText != "", ui.Text(c.supportingText)),
			c.headlineView,
			c.supportingView,
		).Alignment(ui.Leading),
		ui.Spacer(),
		c.trailing,
	).Action(c.action).
		Gap(ui.L16).
		Frame(c.frame).
		Render(ctx)
}

// TList is a composite component (List).
// It displays a vertical collection of rows, optionally with a caption and footer.
// A click handler can be attached to individual entries.
type TList struct {
	caption        core.View     // optional caption above the list
	rows           []core.View   // list entries
	frame          ui.Frame      // layout frame (width/height)
	footer         core.View     // optional footer below the list
	onClickedEntry func(idx int) // handler for row clicks
	colorBody      ui.Color
	colorCaption   ui.Color
	colorFooter    ui.Color
	colorHighlight ui.Color
	colorHover     ui.Color
	onHighlighted  func(idx int) bool
}

// List creates a new TList with the given entries as rows.
func List(entries ...core.View) TList {
	return TList{
		rows:           entries,
		colorFooter:    ui.ColorCardFooter,
		colorBody:      ui.ColorCardBody,
		colorCaption:   ui.ColorCardTop,
		colorHover:     ui.ColorCardFooter,
		colorHighlight: ui.M2,
	}
}

func (c TList) ColorBody(color ui.Color) TList {
	c.colorBody = color
	return c
}

func (c TList) ColorCaption(color ui.Color) TList {
	c.colorCaption = color
	return c
}

func (c TList) ColorHighlight(color ui.Color) TList {
	c.colorHighlight = color
	return c
}

func (c TList) ColorHover(color ui.Color) TList {
	c.colorHover = color
	return c
}

func (c TList) ColorFooter(color ui.Color) TList {
	c.colorFooter = color
	return c
}

// Caption sets an optional caption view above the list.
func (c TList) Caption(s core.View) TList {
	c.caption = s
	return c
}

// Frame sets the layout frame of the list.
func (c TList) Frame(frame ui.Frame) TList {
	c.frame = frame
	return c
}

// FullWidth expands the list to use the full available width.
func (c TList) FullWidth() TList {
	c.frame.Width = ui.Full
	return c
}

// Footer sets an optional footer view below the list.
func (c TList) Footer(s core.View) TList {
	c.footer = s
	return c
}

// OnEntryClicked sets a callback for when a row is clicked.
func (c TList) OnEntryClicked(fn func(idx int)) TList {
	c.onClickedEntry = fn
	return c
}

func (c TList) OnHighlighted(fn func(idx int) bool) TList {
	c.onHighlighted = fn
	return c
}

func (c TList) With(fn func(c TList) TList) TList {
	return fn(c)
}

// Render builds the visual representation of the list.
// It renders an optional caption, the list rows (with separators and
// optional click handling), and an optional footer inside a styled card.
func (c TList) Render(ctx core.RenderContext) core.RenderNode {
	rows := make([]core.View, 0, len(c.rows)*2+3)
	if c.caption != nil {
		rows = append(rows, ui.HStack(c.caption).Alignment(ui.Leading).FullWidth().BackgroundColor(c.colorCaption).Padding(ui.Padding{}.Vertical(ui.L8).Horizontal(ui.L16)))
	}

	for idx, row := range c.rows {
		if row == nil {
			continue
		}

		hstack := ui.HStack(row).HoveredBackgroundColor(c.colorHover)
		if c.onClickedEntry != nil {
			hstack = hstack.Action(func() {
				c.onClickedEntry(idx)
			})
		}

		if fn := c.onHighlighted; fn != nil {
			if fn(idx) {
				hstack = hstack.BackgroundColor(c.colorHover)
			}
		}

		rows = append(rows, hstack.Padding(ui.Padding{}.Vertical(ui.L8).Horizontal(ui.L16)).Frame(ui.Frame{}.FullWidth()))
		if idx < len(c.rows)-1 {
			rows = append(rows, ui.HStack(ui.HLine().Padding(ui.Padding{})).FullWidth().Padding(ui.Padding{}.Horizontal(ui.L16)))
		}
	}

	if c.footer != nil {
		if c.colorFooter == "" && c.colorBody == "" {
			// UX note: if everything is uncolored, introduce an hline for visual separation with footer, otherwise the footer cannot be distinguished
			rows = append(rows, ui.HStack(ui.HLine().Padding(ui.Padding{})).FullWidth().Padding(ui.Padding{}.Horizontal(ui.L16)))
		}
		rows = append(rows, ui.HStack(c.footer).Alignment(ui.Leading).FullWidth().BackgroundColor(c.colorFooter).Padding(ui.Padding{}.Vertical(ui.L16).Horizontal(ui.L16)))
	}

	return ui.VStack(rows...).
		BackgroundColor(c.colorBody).
		Border(ui.Border{}.Radius(ui.L16)).
		Frame(c.frame).
		Render(ctx)
}
