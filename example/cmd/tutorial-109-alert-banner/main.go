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
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

const (
	text = "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua."
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_109")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			stateOK := core.StateOf[bool](wnd, "stateOK").Init(func() bool { return true })
			stateSuccess := core.StateOf[bool](wnd, "stateSuccess").Init(func() bool { return true })
			stateWarning := core.StateOf[bool](wnd, "stateWarning").Init(func() bool { return true })
			stateError := core.StateOf[bool](wnd, "stateError").Init(func() bool { return true })

			return ui.VStack(
				ui.ThemeSwitcher(
					ui.PrimaryButton(nil).Title("Toggle theme"),
				),
				alert.Banner("Banner Information", text).
					Intent(alert.IntentOk).
					Closeable(stateOK).
					Frame(ui.Frame{Width: ui.L880, MaxWidth: ui.Full}),
				alert.Banner("Banner Success", text).
					Intent(alert.IntentSuccess).
					Closeable(stateSuccess).
					Frame(ui.Frame{Width: ui.L880, MaxWidth: ui.Full}),
				alert.Banner("Banner Warning", text).
					Intent(alert.IntentWarning).
					Closeable(stateWarning).
					Frame(ui.Frame{Width: ui.L880, MaxWidth: ui.Full}),
				alert.Banner("Banner Error", text).
					Intent(alert.IntentError).
					Closeable(stateError).
					Frame(ui.Frame{Width: ui.L880, MaxWidth: ui.Full}),
			).Gap(ui.L32).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
