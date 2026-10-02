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
	app.Register("grid-cell", func(wnd core.Window) core.View {
		return Grid(
			GridCell(Text("rows 1-2, column 1")).
				RowStart(1).RowEnd(3).ColStart(1).ColEnd(2).
				BackgroundColor("#FDE2C4").Padding(Padding{}.All(L8)),
			GridCell(Text("row 1, columns 2-3")).
				RowStart(1).RowEnd(2).ColStart(2).ColEnd(4).
				BackgroundColor("#C9E7F8").Padding(Padding{}.All(L8)),
			GridCell(Text("Alignment(Center)")).
				RowStart(2).RowEnd(3).ColStart(2).ColEnd(4).
				Alignment(Center).
				BackgroundColor("#D8F0D2").Padding(Padding{}.All(L8)),
		).Rows(2).
			Columns(3).
			Gap(L8).
			Heights(L80, L80).
			BackgroundColor(ColorCardBody).
			Frame(Frame{Width: L560})
	})
}
