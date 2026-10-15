// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package ui creates the views of nago in a program, which runs on a server, e.g. in goja, like the package
// presentation/ui of nago. Only the server implements its functions, so such a program does not run in a browser.
//
// A view may take functions of the program, e.g. the function, which renders the cells of a sheet. The window calls
// them later, when it renders, and only for the visible cells:
//
//	func View() ui.View {
//		return ui.Sheet("orders", 50000, 12, cell)
//	}
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui/sheet"
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	nsheet "go.wdy.de/nago/presentation/ui/sheet"
)

// View is a view of nago.
type View struct{ ref js.Ref }

func wrap(v core.View) View {
	return View{ref: js.NewRef(v)}
}

// Sheet returns a sheet with rows and columns like a spreadsheet, whose cells cell returns. The window calls cell for
// the visible cells only, also later, when the user scrolls.
//
//gostrict:host
func Sheet(id string, rows, cols int32, cell func(row, col int32) sheet.Cell) View {
	return wrap(nsheet.Sheet(id, int(rows), int(cols), func(row, col int) nsheet.Cell {
		return cellOf(cell(int32(row), int32(col)))
	}))
}
