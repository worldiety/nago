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
		cfg.SetApplicationID("de.worldiety.tutorial_107")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.Stack(
				ui.VStack(
					ui.Stack(
						ui.ThemeSwitcher(
							ui.PrimaryButton(nil).Title("Toggle theme"),
						),
					).Padding(ui.Padding{Bottom: ui.L32}),
					ui.H1("Ich bin eine H1-Überschrift"),
					ui.H2("Ich bin eine H2-Überschrift"),
					ui.H3("Ich bin eine H3-Überschrift"),
					ui.H4("Ich bin eine H4-Überschrift"),
					ui.H5("Ich bin eine H5-Überschrift"),
					ui.H6("Ich bin eine H6-Überschrift"),
				).Alignment(ui.Leading).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{MaxWidth: ui.L1600}),
			).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
