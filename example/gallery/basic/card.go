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
	app.Register("card", func(wnd core.Window) core.View {
		return cardlayout.Card("Storage").
			Body(Text("You use 3.2 GB of your 10 GB.")).
			Footer(PrimaryButton(func() {}).Title("Upgrade")).
			Frame(Frame{Width: L400})
	})
}
