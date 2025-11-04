// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_52")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.SetDecorator(cfg.NewScaffold().
			Login(true).
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			MenuEntry().Icon(icons.SpeakerWave).Forward("/123").Public().
			SubmenuEntry(func(menu *application.SubMenuBuilder) {
				menu.Title("sub menu")
				menu.MenuEntry().Title("first").Action(func(wnd core.Window) {
					fmt.Println("clicked the first entry")
				})
				menu.MenuEntry().Title("second").Forward(".").Public()
				menu.Icon(icons.QuestionMarkCircle)
			}).
			Breakpoint(1000).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			return ui.VStack(
				ui.Text("hello world!"),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())

		}))

	}).Run()
}
