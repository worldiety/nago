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
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/switcher"
)

func init() {
	app.Register("switcher", func(wnd core.Window) core.View {
		pages := []switcher.TSwitcherPage{
			switcher.SwitcherPage("billing", "Billing", icons.Banknotes,
				VStack(
					Text("Billing").Font(HeadlineSmall),
					Text("Invoices, payment methods and your billing address."),
				).Alignment(Leading).Gap(L8),
			),
			switcher.SwitcherPage("favorites", "Favorites", icons.Heart,
				Text("Everything you marked as favorite."),
			),
			switcher.SwitcherPage("documents", "Documents", icons.DocumentText,
				Text("Contracts and other documents."),
			),
		}

		return switcher.Switcher(pages, core.AutoState[string](wnd)).
			Frame(Frame{Width: L880})
	})
}
