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
	app.Register("vstack", func(wnd core.Window) core.View {
		return VStack(
			Text("Short").BackgroundColor("#C9E7F8").Padding(Padding{}.All(L16)),
			Text("A bit longer").BackgroundColor("#FDE2C4").Padding(Padding{}.All(L16)),
			Text("The longest of them all").BackgroundColor("#D8F0D2").Padding(Padding{}.All(L16)),
		).Gap(L8).
			Alignment(Leading).
			BackgroundColor(ColorCardBody).
			Padding(Padding{}.All(L16)).
			Border(Border{}.Radius(L16))
	})
}
