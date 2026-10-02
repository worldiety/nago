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
	app.Register("color", func(wnd core.Window) core.View {
		colors := []Color{M0, M4, M8, A0, I0, SE0, SW0, SG0, "#FA2C7F", Color("#1940BF").WithTransparency(50)}

		return HStack(
			ForEach(colors, func(c Color) core.View {
				return VStack(
					VStack().BackgroundColor(c).
						Frame(Frame{}.Size(L64, L64)).
						Border(Border{}.Radius(L8).Width(L1).Color(ColorLine)),
					Text(string(c)).Font(Small),
				).Gap(L4)
			})...,
		).Gap(L16).Alignment(Top)
	})
}
