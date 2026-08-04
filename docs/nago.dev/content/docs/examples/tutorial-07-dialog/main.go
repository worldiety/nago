// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"slices"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			isPresented := core.AutoState[bool](wnd)

			return VStack(
				ThemeSwitcher(
					PrimaryButton(nil).Title("Toggle theme"),
				),

				VStack(
					SecondaryButton(func() {
						isPresented.Set(true)
					}).Title("Show Dialog"),

					If(isPresented.Get(), Modal(
						Dialog(Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.")).
							Title(Text("Titel")).
							Footer(PrimaryButton(func() {
								isPresented.Set(false)
							}).Title("Schließen")),
					).OnDismissRequest(func() {
						isPresented.Set(false)
					})),
				).Append(
					slices.Collect(func(yield func(core.View) bool) {
						for i := range 50 {
							yield(Text(fmt.Sprintf("Line %d", i)))
						}
					})...,
				).
					Frame(Frame{}.MatchScreen()),
			).Gap(L32).Padding(Padding{}.All(L16)).Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
