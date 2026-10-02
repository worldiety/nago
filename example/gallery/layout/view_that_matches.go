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
	app.Register("view-that-matches", func(wnd core.Window) core.View {
		return ViewThatMatches(wnd,
			SizeClass(core.SizeClassSmall, func() core.View {
				return Text("small screen: one column").BackgroundColor("#FDE2C4").Padding(Padding{}.All(L16))
			}),
			SizeClass(core.SizeClassLarge, func() core.View {
				return HStack(
					Text("large screen:").BackgroundColor("#C9E7F8").Padding(Padding{}.All(L16)),
					Text("two columns").BackgroundColor("#C9E7F8").Padding(Padding{}.All(L16)),
				).Gap(L8)
			}),
		)
	})
}
