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
	app.Register("border", func(wnd core.Window) core.View {
		return HStack(
			Text("Width + Color").Padding(Padding{}.All(L16)).
				Border(Border{}.Width(L2).Color("#1940BF")),
			Text("Radius").Padding(Padding{}.All(L16)).
				Border(Border{}.Width(L2).Color("#1940BF").Radius(L16)),
			Text("TopRadius").Padding(Padding{}.All(L16)).
				Border(Border{}.Width(L2).Color("#1940BF").TopRadius(L16)),
			Text("Elevate(4)").Padding(Padding{}.All(L16)).
				Border(Border{}.Radius(L8).Elevate(4)),
			VStack(Text("Circle")).BackgroundColor("#C9E7F8").Frame(Frame{}.Size(L80, L80)).
				Border(Border{}.Circle()),
		).Gap(L24)
	})
}
