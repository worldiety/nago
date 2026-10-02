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
	app.Register("modal", func(wnd core.Window) core.View {
		presented := core.AutoState[bool](wnd)

		return VStack(
			PrimaryButton(func() { presented.Set(true) }).Title("Show hint"),
			If(presented.Get(), Overlay(
				VStack(
					Text("Tip of the day").Font(SubTitle),
					Text("Press Escape to close dialogs."),
					SecondaryButton(func() { presented.Set(false) }).Title("Got it"),
				).Alignment(Leading).
					Gap(L8).
					BackgroundColor(M1).
					Border(Border{}.Radius(L16).Shadow(L8)).
					Padding(Padding{}.All(L16)),
			).Top(L24).Right(L24)),
		)
	})
}
