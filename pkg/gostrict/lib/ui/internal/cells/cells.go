// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package cells is a program, which renders a large sheet, for the test and benchmark of callbacks.
package cells

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/sheet"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

const cols = 30

var header = sheet.Style{
	BackgroundColor: "#eeeeee",
	Bold:            true,
	Borders:         sheet.All(sheet.Line{Color: "#999999", Style: sheet.LineSolid}),
}

// Cell returns the header in the first row and column and numbers elsewhere.
func Cell(row, col int32) sheet.Cell {
	if row == 0 || col == 0 {
		return sheet.Cell{Text: text.FromInt(row + col), Style: header}
	}
	return sheet.Cell{Text: text.FromInt(row*cols + col), Value: text.FromInt(row*cols + col), Type: sheet.ValueNumber, Editor: sheet.EditorNumber}
}

// View returns the sheet.
func View() ui.View {
	return ui.Sheet("big", 50000, cols, Cell)
}
