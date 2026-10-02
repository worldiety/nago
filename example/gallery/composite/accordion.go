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
	"go.wdy.de/nago/presentation/ui/accordion"
)

func init() {
	app.Register("accordion", func(wnd core.Window) core.View {
		shipping := core.AutoState[bool](wnd).Init(func() bool { return true })
		returns := core.AutoState[bool](wnd)
		payment := core.AutoState[bool](wnd)

		return VStack(
			accordion.Accordion(
				Text("How long does shipping take?"),
				Text("Orders are shipped within two working days."),
				shipping,
			).FullWidth(),
			accordion.Accordion(
				Text("Can I return an item?"),
				Text("You can return every item within 30 days."),
				returns,
			).FullWidth(),
			accordion.Accordion(
				Text("Which payment methods are available?"),
				Text("Invoice, credit card and direct debit."),
				payment,
			).FullWidth(),
		).Frame(Frame{Width: L480})
	})
}
