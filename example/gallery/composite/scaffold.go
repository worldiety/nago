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
	icons "go.wdy.de/nago/presentation/icons/hero/outline"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("scaffold", func(wnd core.Window) core.View {
		return Scaffold(ScaffoldAlignmentLeading).
			Logo(Image().Embed(icons.CubeTransparent).Frame(Frame{}.Size(L48, L48))).
			Menu(
				ForwardScaffoldMenuEntry(wnd, icons.Home, "Dashboard", "scaffold"),
				ForwardScaffoldMenuEntry(wnd, icons.Users, "Customers", "customers"),
				ParentScaffoldMenuEntry(wnd, icons.ChartBar, "Reports",
					ScaffoldMenuEntry{Title: "Sales"},
					ScaffoldMenuEntry{Title: "Inventory"},
				),
			).
			Body(VStack(
				Text("Dashboard").Font(Title),
				Text("Welcome back, Ada."),
			).Alignment(Leading).Gap(L16).Padding(Padding{}.All(L32)))
	})
}
