// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("table", func(wnd core.Window) core.View {
		type order struct {
			ID, Customer, Status, Total string
		}

		orders := []order{
			{"#1042", "Ada Lovelace", "Shipped", "129.00 €"},
			{"#1043", "Grace Hopper", "Processing", "48.50 €"},
			{"#1044", "Alan Turing", "Delivered", "312.20 €"},
		}

		return Table(
			TableColumn(Text("Order")),
			TableColumn(Text("Customer")),
			TableColumn(Text("Status")),
			TableColumn(Text("Total")).Alignment(Trailing),
		).Rows(
			ForEach(orders, func(o order) TTableRow {
				return TableRow(
					TableCell(Text(o.ID)),
					TableCell(Text(o.Customer)),
					TableCell(Text(o.Status)),
					TableCell(Text(o.Total)).Alignment(Trailing),
				).Action(func() {
					// open the order
				}).HoveredBackgroundColor(ColorCardFooter)
			})...,
		).
			CellPadding(Padding{}.Horizontal(L16).Vertical(L12)).
			Border(Border{}.Radius(L16)).
			Frame(Frame{Width: L560})
	})
}
