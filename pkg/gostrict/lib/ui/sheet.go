// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui/sheet"
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nsheet "go.wdy.de/nago/presentation/ui/sheet"
)

// TSheet is a sheet like a spreadsheet, see Sheet.
type TSheet struct{ ref js.Ref }

func (TSheet) isView() {}

// Sheet returns a sheet with rows and columns, whose cells cell returns, like sheet.Sheet of nago. The window calls
// cell for the visible cells only, also later, when the user scrolls.
//
//gostrict:host
func Sheet(id string, rows, cols int32, cell func(row, col int32) sheet.Cell) TSheet {
	return TSheet{ref: wrap(nsheet.Sheet(id, int(rows), int(cols), func(row, col int) nsheet.Cell {
		return cellOf(cell(int32(row), int32(col)))
	}))}
}

// Frozen keeps the leading rows and columns visible, while the user scrolls, e.g. a header.
//
//gostrict:host
func (c TSheet) Frozen(rows, cols int32) TSheet {
	return TSheet{ref: wrap(get[nsheet.TSheet](c.ref).Frozen(int(rows), int(cols)))}
}

// ColumnWidth sets the width of the columns in dp.
//
//gostrict:host
func (c TSheet) ColumnWidth(width float64) TSheet {
	return TSheet{ref: wrap(get[nsheet.TSheet](c.ref).ColumnWidth(width))}
}

// Disabled disables the editors of the cells.
//
//gostrict:host
func (c TSheet) Disabled(disabled bool) TSheet {
	return TSheet{ref: wrap(get[nsheet.TSheet](c.ref).Disabled(disabled))}
}

// OnEdit calls fn with each value, which the user enters into a cell with an editor. An error is shown at the
// cell.
//
//gostrict:host
func (c TSheet) OnEdit(fn func(e sheet.Edit) error) TSheet {
	return TSheet{ref: wrap(get[nsheet.TSheet](c.ref).OnEdit(func(e nsheet.Edit) error {
		return fn(sheet.Edit{Row: int32(e.Row), Col: int32(e.Col), Value: e.Value})
	}))}
}
