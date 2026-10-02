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
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

func init() {
	app.Register("card-layout", func(wnd core.Window) core.View {
		return cardlayout.Layout(
			cardlayout.Card("Revenue").Body(Text("12,400 €")),
			cardlayout.Card("Orders").Body(Text("318")).Footer(Text("+12 % this week")),
			cardlayout.Card("Customers").Body(Text("1,024")),
		).Columns(core.SizeClassLarge, 3).
			Frame(Frame{Width: L880})
	})
}
