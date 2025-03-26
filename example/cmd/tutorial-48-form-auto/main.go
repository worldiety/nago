// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/web/vuejs"
)

type SomeThing struct {
	Name string `id:"abc1234"`
	When xtime.Date
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			thingState := core.AutoState[SomeThing](wnd)
			return ui.VStack(
				form.Auto(form.AutoOptions{}, thingState),
				ui.PrimaryButton(func() {
					fmt.Printf("Thing: %v\n", thingState.Get())
				}).Title("print"),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())
		}))

	}).Run()
}
