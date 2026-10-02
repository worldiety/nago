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
	app.Register("shadow", func(wnd core.Window) core.View {
		return HStack(
			Text("Shadow(L8)").Padding(Padding{}.All(L24)).
				Border(Border{}.Radius(L8).Shadow(L8)),
			Text("custom Shadow").Padding(Padding{}.All(L24)).
				Border(Border{
					TopLeftRadius:     L8,
					TopRightRadius:    L8,
					BottomLeftRadius:  L8,
					BottomRightRadius: L8,
					BoxShadow: Shadow{
						Color:  "#FA2C7F80",
						Radius: L16,
						X:      L8,
						Y:      L8,
					},
				}),
		).Gap(L48).Padding(Padding{}.All(L32))
	})
}
