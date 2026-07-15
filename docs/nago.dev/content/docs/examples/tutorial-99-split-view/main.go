// Copyright (c) 2025 worldiety GmbH
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
		cfg.SetApplicationID("de.worldiety.tutorial_99")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			state := core.AutoState[float64](wnd).Init(func() float64 {
				return 0.7
			})

			return ui.VStack(
				ui.Stack(
					ui.PrimaryButton(func() {
						if wnd.Info().PrefersLight() {
							wnd.SetColorScheme(core.Dark)
						} else {
							wnd.SetColorScheme(core.Light)
						}
					}).Title("Toggle theme"),
					ui.PrimaryButton(func() {
						state.Set(0.5)
					}).Title("Center"),
				).Alignment(ui.Center).
					Gap(ui.L8),
				ui.Stack(
					ui.SplitView(
						ui.ScrollView(
							ui.Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."),
						).BackgroundColor(ui.M2).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{Height: ui.Full}),
						ui.ScrollView(
							ui.Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."),
						).BackgroundColor(ui.M3).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{Height: ui.Full}),
					).
						InputValue(state).
						MinRatio(0.1).
						MaxRatio(0.9).
						Frame(ui.Frame{Height: ui.Full}),
				).Frame(ui.Frame{MaxWidth: ui.L880, Height: ui.L480}),
			).Gap(ui.L32).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
