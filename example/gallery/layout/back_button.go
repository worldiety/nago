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
	"go.wdy.de/nago/presentation/ui/layout"
)

func init() {
	app.Register("back-button", func(wnd core.Window) core.View {
		return VStack(
			layout.WithBackButton(wnd, VStack(
				Text("Order #4711").Font(HeadlineSmall),
				Text("Shipped on 2 October."),
			).Alignment(Leading).FullWidth()),
		).Frame(Frame{Width: L480})
	})
}
