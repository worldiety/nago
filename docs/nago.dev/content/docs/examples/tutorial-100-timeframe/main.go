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
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timeframe"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_100")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			//stateDay := core.StateOf[xtime.Date](wnd, "stateDay")
			//stateStart := core.StateOf[time.Duration](wnd, "stateStart")
			//stateEnd := core.StateOf[time.Duration](wnd, "stateEnd")
			stateTarget := core.StateOf[xtime.TimeFrame](wnd, "stateTarget")

			return ui.VStack(
				ui.PrimaryButton(func() {
					if wnd.Info().ColorScheme == core.Light {
						wnd.SetColorScheme(core.Dark)
					} else {
						wnd.SetColorScheme(core.Light)
					}
				}).Title("Toggle theme"),
				ui.VStack(
					timeframe.Picker("Standard", stateTarget),
					timeframe.Picker("Disabled", stateTarget).Disabled(true),
				).Gap(ui.L32),
			).Gap(ui.L32).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
