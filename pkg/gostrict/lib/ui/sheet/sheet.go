// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package sheet declares the cells of a sheet, which ui.Sheet shows, like the package presentation/ui/sheet of nago.
// A cell is a record, which the function of the program returns for each visible cell:
//
//	var header = sheet.Style{BackgroundColor: "#eeeeee", Bold: true}
//
//	func cell(row, col int32) sheet.Cell {
//		if row == 0 {
//			return sheet.Cell{Text: "Spalte " + text.FromInt(col), Style: header}
//		}
//		return sheet.Cell{Text: text.FromInt(row * col), Type: sheet.ValueNumber}
//	}
//
// A style in a package variable is shared by all its cells, so it costs nothing per cell.
package sheet

// ValueType tells, how a value is aligned by default and how it is copied into a spreadsheet application.
type ValueType int32

const (
	ValueText ValueType = iota
	// ValueNumber has a [Cell.Value] in machine notation, e.g. 1234.5, while [Cell.Text] may be formatted.
	ValueNumber
	ValueBool
	// ValueDate has a [Cell.Value] in ISO 8601 notation, e.g. 2026-10-14.
	ValueDate
)

// Editor is the editor of a cell.
type Editor int32

const (
	EditorNone Editor = iota
	EditorText
	EditorNumber
	// EditorCheckbox toggles the value between "true" and "false".
	EditorCheckbox
)

// Align aligns the content of a cell horizontally. AlignAuto aligns numbers at the end and anything else at the
// start.
type Align int32

const (
	AlignAuto Align = iota
	AlignStart
	AlignCenter
	AlignEnd
)

// Wrap tells, what happens with text, which does not fit into a cell.
type Wrap int32

const (
	WrapClip Wrap = iota
	WrapEllipsis
	// WrapWrap breaks the text into lines, as many as fit into the height of the row.
	WrapWrap
)

// LineStyle is the style of a border line.
type LineStyle int32

const (
	LineNone LineStyle = iota
	LineSolid
	LineDashed
	LineDotted
	LineDouble
)

// Line is a border at one side of a cell. A zero line is the grid line of the sheet.
type Line struct {
	// Color is a color like "#ff0000" or a color of the theme like "M0".
	Color string
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
	BackgroundColor string
	Color           string
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
	// Editor lets the user change the cell. The editor is shown with the Value, or the Text if the Value is empty.
	Editor Editor
}

// Edit is a value, which the user has entered into a cell.
type Edit struct {
	Row, Col int32
	Value    string
}
