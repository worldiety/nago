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

// the demo is shared by the table cell, table column and table row pages
func init() {
	app.Register("table-parts", func(wnd core.Window) core.View {
		return Table(
			TableColumn(Text("Name")),
			TableColumn(Text("Role")).Width(L160),
			TableColumn(Text("Status")).Alignment(Trailing),
		).Rows(
			TableRow(
				TableCell(Text("Ada")),
				TableCell(Text("Admin")),
				TableCell(Text("active")),
			),
			TableRow(
				TableCell(Text("Linus")),
				TableCell(Text("Editor")),
				TableCell(Text("invited")).BackgroundColor("#FDE2C4"),
			).HoveredBackgroundColor(ColorCardFooter),
			TableRow(
				TableCell(Text("Grace: ColSpan(2)")).ColSpan(2),
				TableCell(Text("active")),
			).BackgroundColor("#C9E7F8"),
		).Frame(Frame{Width: L560})
	})
}
