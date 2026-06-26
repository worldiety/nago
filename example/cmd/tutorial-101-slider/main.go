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
			stateDefault := core.StateOf[float64](wnd, "stateDefault").Init(func() float64 { return 3 })
			stateMarkers := core.StateOf[float64](wnd, "stateMarkers").Init(func() float64 { return 3 })
			stateUnit := core.StateOf[float64](wnd, "stateUnit").Init(func() float64 { return 3 })
			stateSupport := core.StateOf[float64](wnd, "stateSupport").Init(func() float64 { return 3 })
			stateError := core.StateOf[float64](wnd, "stateError").Init(func() float64 { return 3 })
			stateSmallSteps := core.StateOf[float64](wnd, "stateSmallSteps").Init(func() float64 { return 3 })
			stateDisabled := core.StateOf[float64](wnd, "stateDisabled").Init(func() float64 { return 3 })

			return ui.VStack(
				ui.PrimaryButton(func() {
					if wnd.Info().ColorScheme == core.Light {
						wnd.SetColorScheme(core.Dark)
					} else {
						wnd.SetColorScheme(core.Light)
					}
				}).Title("Toggle theme"),
				ui.Slider(0, 10).Label("Standard").Frame(ui.Frame{MinWidth: ui.L200}).InputValue(stateDefault),
				ui.Slider(0, 10).Label("Mit Marker").Frame(ui.Frame{MinWidth: ui.L200}).InputValue(stateMarkers).ShowMarkers(true),
				ui.Slider(0, 10).Label("Mit Einheit").Frame(ui.Frame{MinWidth: ui.L200}).InputValue(stateUnit).Unit("g"),
				ui.Slider(0, 10).Label("Mit Support").Frame(ui.Frame{MinWidth: ui.L200}).InputValue(stateSupport).SupportingText("Ich bin ein Support-Text"),
				ui.Slider(0, 10).Label("Mit Fehler").Frame(ui.Frame{MinWidth: ui.L200}).InputValue(stateError).ErrorText("Ich bin ein Fehler-Text"),
				ui.Slider(0, 10).Label("Feingranular").Frame(ui.Frame{MinWidth: ui.L200}).InputValue(stateSmallSteps).Step(0.01),
				ui.Slider(0, 10).Label("Disabled").Frame(ui.Frame{MinWidth: ui.L200}).InputValue(stateDisabled).Disabled(true),
			).Gap(ui.L32).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
