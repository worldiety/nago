// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_101")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			stateDefault := core.StateOf[float64](wnd, "state").Init(func() float64 { return 3 })

			return ui.VStack(
				ui.PrimaryButton(func() {
					if wnd.Info().ColorScheme == core.Light {
						wnd.SetColorScheme(core.Dark)
					} else {
						wnd.SetColorScheme(core.Light)
					}
				}).Title("Toggle theme"),
				ui.Slider(0, 10).Label("Standard").SupportingText("Ich bin ein Support-Text").InputValue(stateDefault).ShowMarkers(true),
			).Gap(ui.L32).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
