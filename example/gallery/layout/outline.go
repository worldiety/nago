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
	app.Register("outline", func(wnd core.Window) core.View {
		return HStack(
			VStack(Text("Outline")).
				Outline(Outline{Style: OutlineSolid, Width: 2, Offset: 4, Color: "#FA2C7F"}).
				BackgroundColor("#C9E7F8").
				Padding(Padding{}.All(L16)),
			VStack(Text("Inside()")).
				Outline(Outline{Style: OutlineSolid, Width: 2, Color: "#FA2C7F"}.Inside()).
				BackgroundColor("#C9E7F8").
				Padding(Padding{}.All(L16)),
		).Gap(L32).Padding(Padding{}.All(L16))
	})
}
