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
	app.Register("divider", func(wnd core.Window) core.View {
		return VStack(
			Text("Above"),
			HLine(),
			Text("Below"),
			HLineWithColor(ColorAccent),
			HStack(
				Text("Left"),
				VLine(),
				Text("Right"),
			).Gap(L16).Frame(Frame{Height: L48}),
		).Alignment(Leading).Frame(Frame{Width: L320})
	})
}
