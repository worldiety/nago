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
	app.Register("box", func(wnd core.Window) core.View {
		return HStack(
			Box(BoxLayout{
				TopLeading:     Text("TopLeading"),
				Center:         Text("Center"),
				BottomTrailing: Text("BottomTrailing"),
			}).BackgroundColor("#C9E7F8").
				Padding(Padding{}.All(L8)).
				Frame(Frame{}.Size(L320, L160)),

			BoxAlign(BottomTrailing, Text("BoxAlign")).
				BackgroundColor("#FDE2C4").
				Padding(Padding{}.All(L8)).
				Frame(Frame{}.Size(L200, L160)),
		).Gap(L16)
	})
}
