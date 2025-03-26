// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(
				Text("A simple grid row, by default cells are spanned completely"),
				Grid(
					GridCell(Text("cell 1").BackgroundColor("#ff0000")),
					GridCell(Text("cell 2").BackgroundColor("#ff0000")),
					GridCell(Text("cell 3").BackgroundColor("#ff0000")),
				).
					Rows(1).
					BackgroundColor("#00ff00").
					Frame(Frame{}.Size(L320, L320)),

				Text("A simple grid row using stacks for alignment"),
				Grid(
					GridCell(VStack(Text("cell 1").BackgroundColor("#ff0000")).Alignment(Leading)),
					GridCell(VStack(Text("cell 2").BackgroundColor("#ff0000")).Alignment(Center)),
					GridCell(VStack(Text("cell 3").BackgroundColor("#ff0000")).Alignment(Trailing)),
				).
					Rows(1).
					BackgroundColor("#00ff00").
					Frame(Frame{}.Size(L320, L320)),

				Text("Cell alignment rules"),
				Grid(
					GridCell(Text("Leading").BackgroundColor("#ff0000")).
						Alignment(Leading),
					GridCell(Text("Center").BackgroundColor("#ff0000")).
						Alignment(Center),
					GridCell(Text("Trailing").BackgroundColor("#ff0000")).
						Alignment(Trailing),
					GridCell(Text("Top").BackgroundColor("#ff0000")).
						Alignment(Top),
					GridCell(Text("Bottom").BackgroundColor("#ff0000")).
						Alignment(Bottom),
					GridCell(Text("TopLeading").BackgroundColor("#ff0000")).
						Alignment(TopLeading),
					GridCell(Text("TopTrailing").BackgroundColor("#ff0000")).
						Alignment(TopTrailing),
					GridCell(Text("BottomLeading").BackgroundColor("#ff0000")).
						Alignment(BottomLeading),
					GridCell(Text("BottomTrailing").BackgroundColor("#ff0000")).
						Alignment(BottomTrailing),
				).
					Rows(1).
					BackgroundColor("#00ff00").
					Frame(Frame{}.Size("100%", L320)),
			).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
