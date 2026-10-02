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
	app.Register("padding", func(wnd core.Window) core.View {
		return HStack(
			VStack(Text("All(L16)").BackgroundColor("#C9E7F8")).
				BackgroundColor("#FDE2C4").
				Padding(Padding{}.All(L16)),
			VStack(Text("Horizontal(L32)").BackgroundColor("#C9E7F8")).
				BackgroundColor("#FDE2C4").
				Padding(Padding{}.Horizontal(L32)),
			VStack(Text("Vertical(L32)").BackgroundColor("#C9E7F8")).
				BackgroundColor("#FDE2C4").
				Padding(Padding{}.Vertical(L32)),
			VStack(Text("Padding{Left: L48}").BackgroundColor("#C9E7F8")).
				BackgroundColor("#FDE2C4").
				Padding(Padding{Left: L48}),
		).Gap(L16).Alignment(Top)
	})
}
