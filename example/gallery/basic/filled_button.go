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
)

func init() {
	app.Register("filled-button", func(wnd core.Window) core.View {
		return HStack(
			FilledButton(SG0, func() {}).Title("Approve").TextColor(ColorWhite),
			FilledButton(SE0, func() {}).Title("Reject").TextColor(ColorWhite).PreIcon(icons.XMark),
		).Gap(L16)
	})
}
