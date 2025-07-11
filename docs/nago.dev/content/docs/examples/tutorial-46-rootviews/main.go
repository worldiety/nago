// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			return ui.VStack(
				ui.PrimaryButton(func() {
					wnd.Navigation().ForwardTo("hello", core.Values{"msg": "world"})
				}).Title("go to hello"),
			).Frame(ui.Frame{}.MatchScreen())
		}))

		cfg.RootView("hello", cfg.DecorateRootView(func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("your message is hello "+wnd.Values()["msg"]),
				ui.PrimaryButton(func() {
					wnd.Navigation().Back()
				}).Title("back"),
			).Frame(ui.Frame{}.MatchScreen())
		}))

	}).Run()
}
