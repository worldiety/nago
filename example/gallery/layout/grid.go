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
	app.Register("grid", func(wnd core.Window) core.View {
		return Grid(
			GridCell(Text("ColSpan(3)")).ColSpan(3).BackgroundColor("#FDE2C4").Padding(Padding{}.All(L8)),
			GridCell(Text("A")).BackgroundColor("#C9E7F8").Padding(Padding{}.All(L8)),
			GridCell(Text("B")).BackgroundColor("#C9E7F8").Padding(Padding{}.All(L8)),
			GridCell(Text("C")).BackgroundColor("#C9E7F8").Padding(Padding{}.All(L8)),
			GridCell(Text("D")).BackgroundColor("#D8F0D2").Padding(Padding{}.All(L8)),
			GridCell(Text("E")).BackgroundColor("#D8F0D2").Padding(Padding{}.All(L8)),
			GridCell(Text("F")).BackgroundColor("#D8F0D2").Padding(Padding{}.All(L8)),
		).Columns(3).
			Gap(L8).
			Widths(L120, "1fr", L120).
			Frame(Frame{Width: L480})
	})
}
