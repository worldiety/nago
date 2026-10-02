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
	app.Register("length", func(wnd core.Window) core.View {
		lengths := []Length{L16, L48, L120, L200, L320, Relative(0.5), Absolute(100), "calc(100% - 4rem)"}

		return VStack(
			ForEach(lengths, func(l Length) core.View {
				return HStack(
					Text(string(l)).Frame(Frame{Width: L200}),
					VStack().BackgroundColor("#1CCDFB").Frame(Frame{Width: l, Height: L16}),
				).Alignment(Leading).Frame(Frame{Width: L560})
			})...,
		).Alignment(Leading).Gap(L8)
	})
}
