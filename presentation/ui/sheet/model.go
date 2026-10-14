// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package sheet provides a spreadsheet like table for large data. Only the cells of a window around the visible
// area are rendered, so a sheet with tens of thousands of rows renders as fast as a small one. The data is never
// kept in a state: the application provides a function which returns a cell on demand, and an edit is an event,
// which the application handles with its own use cases. See [Sheet].
//
// Every change of the window, e.g. while the user scrolls, renders the whole page again, like any other change of
// a state. On a page with few other views, scrolling is as fluent as in a spreadsheet application. Many other views
// slow it down in the browser, roughly 0.06 ms per DOM node: with 500 further rows of a form beside the sheet, a
// jump takes about 100 ms, with 2000 about 500 ms, and placeholders show while scrolling. So a large sheet should
// rather get a page of its own. See TestHeavyPageScroll to measure it.
package sheet

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// ValueType tells, how a value is aligned by default and how it is copied into a spreadsheet application.
type ValueType int

const (
	ValueText ValueType = iota
	// ValueNumber has a [Cell.Value] in machine notation, e.g. 1234.5, while [Cell.Text] may be formatted.
	ValueNumber
	ValueBool
	// ValueDate has a [Cell.Value] in ISO 8601 notation, e.g. 2026-10-14.
	ValueDate
)

// Editor is the editor of a cell. Every edit is passed to [TSheet.OnEdit].
type Editor int

const (
	EditorNone Editor = iota
	EditorText
	EditorNumber
	// EditorSelect offers the [Cell.Options].
	EditorSelect
	// EditorCheckbox toggles the value between "true" and "false".
	EditorCheckbox
)

// Align aligns the content of a cell horizontally. AlignAuto aligns numbers at the end and anything else at the
// start.
type Align int

const (
	AlignAuto Align = iota
	AlignStart
	AlignCenter
	AlignEnd
)

// Wrap tells, what happens with text, which does not fit into a cell.
type Wrap int

const (
	WrapClip Wrap = iota
	WrapEllipsis
	// WrapWrap breaks the text into lines, as many as fit into the height of the row.
	WrapWrap
)

// LineStyle is the style of a border line.
type LineStyle int

const (
	LineNone LineStyle = iota
	LineSolid
	LineDashed
	LineDotted
	LineDouble
)

// Line is a border at one side of a cell. A zero line is the grid line of the sheet.
type Line struct {
	Color ui.Color
	// Width in dp, 1 if zero.
	Width float64
	Style LineStyle
}

// Borders of a cell.
type Borders struct {
	Top, Right, Bottom, Left Line
}

// All returns borders with the same line at every side.
func All(line Line) Borders {
	return Borders{Top: line, Right: line, Bottom: line, Left: line}
}

// Style of a cell. Equal styles are transferred once, so a sheet with many cells of few styles stays small.
type Style struct {
	BackgroundColor ui.Color
	Color           ui.Color
	Bold            bool
	Italic          bool
	Underline       bool
	Align           Align
	Wrap            Wrap
	Borders         Borders
	// PaddingX is the horizontal padding in dp, the default of the sheet if zero.
	PaddingX float64
	// FontSize in dp, the default if zero.
	FontSize float64
}

// Cell is the content of a cell.
type Cell struct {
	// Text is shown, e.g. a formatted number like "1.234,50 €".
	Text string
	// Value is copied, e.g. into a spreadsheet application, see [ValueType]. If empty, Text is copied.
	Value string
	Type  ValueType
	Style Style
	// Editor lets the user change the cell, see [TSheet.OnEdit]. The editor is shown with the Value, or the
	// Text if the Value is empty.
	Editor  Editor
	Options []string

	// View is shown instead of the Text, e.g. a button or an avatar. It is rendered for the visible cells only, but
	// still costs more than a text, so it should be used for a few columns. Text and Value are copied. The view
	// must fit into the height of the row, so a compact view like a text with an action fits better than a
	// button, which needs a higher row, see [TSheet.RowHeights] and [TSheet.Density]. If an action of the view
	// changes the model of the application but no state, a state must be set as well, so that the window renders
	// again. An edit with [TSheet.OnEdit] renders again anyway.
	View core.View
}

// Merge joins a rectangle of cells into the cell at its top left, which is shown across the rectangle.
type Merge struct {
	Row, Col   int
	Rows, Cols int
}

func (m Merge) contains(row, col int) bool {
	return row >= m.Row && row < m.Row+m.Rows && col >= m.Col && col < m.Col+m.Cols
}

// Layout is what the user has changed of the layout, see [TSheet.Layout]. It complements the layout of the
// application: the user may change the width of a column, hide rows and columns and show them again, but not the
// ones, which the application hides.
type Layout struct {
	ColWidths  map[int]float64 `json:"colWidths,omitempty"`
	HiddenRows []int           `json:"hiddenRows,omitempty"`
	HiddenCols []int           `json:"hiddenCols,omitempty"`
}

// Edit is a value, which the user has entered into a cell.
type Edit struct {
	Row, Col int
	Value    string
}

// Range is a rectangle of cells.
type Range struct {
	Row, Col   int
	Rows, Cols int
}

// Density defines the height of the rows and the padding of the cells.
type Density int

const (
	// DensityCompact is denser than a [ui.Table] and close to a spreadsheet application.
	DensityCompact Density = iota
	DensityDefault
	DensityComfortable
)

func (d Density) rowHeight() float64 {
	switch d {
	case DensityDefault:
		return 32
	case DensityComfortable:
		return 40
	default:
		return 24
	}
}
