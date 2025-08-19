// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cardlayout

import (
	"slices"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// sizeClassColumns defines a responsive rule for card layouts.
// It maps a given WindowSizeClass to a specific number of columns.
type sizeClassColumns struct {
	SizeClass core.WindowSizeClass
	Columns   int
}

// TCardLayout is a container component (Card Layout).
// It organizes child views into a grid-like layout, typically using multiple
// columns. The number of columns can be defined globally or customized per
// window size class to enable responsive design.
type TCardLayout struct {
	children       []core.View
	frame          ui.Frame
	padding        ui.Padding
	customColumns  []sizeClassColumns
	defaultColumns int
}

// Layout creates a new TCardLayout with the given child views.
// Nil children are automatically skipped. By default, the layout uses 3 columns.
func Layout(children ...core.View) TCardLayout {
	tmp := make([]core.View, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}

		tmp = append(tmp, child)
	}

	return TCardLayout{children: tmp, defaultColumns: 3}
}

// Columns sets a custom number of columns for a specific WindowSizeClass.
// This allows the layout to adapt responsively to different screen sizes.
func (c TCardLayout) Columns(class core.WindowSizeClass, columns int) TCardLayout {
	c.customColumns = append(c.customColumns, sizeClassColumns{
		SizeClass: class,
		Columns:   columns,
	})

	return c
}

// Frame sets the frame (size and positioning) for the card layout.
func (c TCardLayout) Frame(frame ui.Frame) TCardLayout {
	c.frame = frame
	return c
}

// Padding sets the inner spacing for the card layout.
// This allows customizing the distance between the layout border and its content.
func (c TCardLayout) Padding(padding ui.Padding) TCardLayout {
	c.padding = padding
	return c
}

// Render builds the card layout as a responsive grid.
// It determines the number of columns based on either:
//   - explicitly defined rules for specific window size classes, or
//   - default fallbacks (1 column for small/medium, 2 for large, 3 by default).
//
// Each child is rendered into a GridCell, with consistent spacing, padding,
// and frame configuration applied to the overall layout.
func (c TCardLayout) Render(ctx core.RenderContext) core.RenderNode {
	columns := c.defaultColumns
	wnd := ctx.Window()
	if len(c.customColumns) > 0 {
		for _, column := range c.customColumns {
			if column.SizeClass == wnd.Info().SizeClass {
				columns = column.Columns
				break
			}
		}
	} else {
		switch wnd.Info().SizeClass {
		case core.SizeClassSmall:
			columns = 1
		case core.SizeClassMedium:
			columns = 1
		case core.SizeClassLarge:
			columns = 2
		}
	}

	return ui.Grid(
		ui.Each(slices.Values(c.children), func(t core.View) ui.TGridCell {
			return ui.GridCell(t)
		})...,
	).Gap(ui.L16).
		Columns(columns).
		Frame(c.frame).
		Padding(c.padding).
		Render(ctx)
}

type TitleStyle int

const (
	TitleLarge TitleStyle = iota
	TitleCompact
)

// TCard is a basic component (Card).
// It represents a structured UI element that can display content in three
// sections: a title, a body, and an optional footer.
// Cards are typically used to group related information or actions together
// in a visually distinct block, and they support custom styling, padding,
// and layout adjustments via frame and title style.
type TCard struct {
	title         string
	body          core.View
	footer        core.View
	frame         ui.Frame
	style         TitleStyle
	padding       ui.Padding
	customPadding bool
}

// Card creates a new card with a title and default padding.
func Card(title string) TCard {
	return TCard{title: title, padding: ui.Padding{Right: ui.L40, Left: ui.L40, Bottom: ui.L40, Top: ""}}
}

// Style sets the title style of the card (e.g., heading level or visual variant).
func (c TCard) Style(style TitleStyle) TCard {
	c.style = style
	return c
}

// Body defines the main content area of the card.
func (c TCard) Body(view core.View) TCard {
	c.body = view
	return c
}

// Padding overrides the default padding of the card and marks it as custom.
func (c TCard) Padding(padding ui.Padding) TCard {
	c.customPadding = true
	c.padding = padding
	return c
}

// Footer adds a footer view below the card body, typically for actions or secondary info.
func (c TCard) Footer(view core.View) TCard {
	c.footer = view
	return c
}

// Frame sets the layout frame (size and positioning) of the card.
func (c TCard) Frame(frame ui.Frame) TCard {
	c.frame = frame
	return c
}

// Render builds and displays the card with title, body, optional footer, padding, and styling.
func (c TCard) Render(ctx core.RenderContext) core.RenderNode {
	var bodyTopPadding ui.Length
	var title core.View
	if c.title != "" {
		switch c.style {
		case TitleCompact:
			bodyTopPadding = ui.L16
			title = ui.HStack(
				ui.Text(c.title),
			).
				Alignment(ui.Leading).
				BackgroundColor(ui.ColorCardTop).
				Padding(ui.Padding{}.All(ui.L8)).
				Frame(ui.Frame{Height: c.padding.Left}.FullWidth())
		default:
			title = ui.VStack(
				ui.Text(c.title).Font(ui.Title),
				ui.HLineWithColor(ui.ColorAccent),
			).Padding(ui.Padding{Top: ui.L40, Left: c.padding.Left, Right: c.padding.Right})
		}
	}

	padding := c.padding
	if !c.customPadding {
		padding.Top = bodyTopPadding
		if padding.Top == "" && c.title == "" {
			padding.Top = padding.Left
		}
	}

	var border ui.Border
	if padding.Left != "" || padding.Bottom != "" {
		border = ui.Border{}.Radius(ui.L16)
	}

	return ui.VStack(
		title,
		ui.VStack(c.body).
			Alignment(ui.Leading).
			FullWidth().
			Padding(padding),
		ui.Spacer(),
		ui.If(c.footer != nil, ui.HStack(c.footer).
			FullWidth().
			Alignment(ui.Trailing).
			BackgroundColor(ui.ColorCardFooter).
			Padding(ui.Padding{}.All(ui.L12))),
	).Alignment(ui.Leading).
		BackgroundColor(ui.ColorCardBody).
		Border(border).
		Frame(c.frame).
		Render(ctx)
}
