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
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

type SomeThing struct {
	Name   string `id:"abc1234"`
	When   xtime.Date
	Who    user.ID   `source:"nago.users"`
	Who2   user.ID   `source:"nago.users"`
	Others []user.ID `source:"nago.users"`
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		option.MustZero(cfg.StandardSystems())
		uid := option.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			thingState := core.AutoState[SomeThing](wnd).Init(func() SomeThing {
				return SomeThing{
					Who:    uid,
					Others: []user.ID{uid},
				}
			})
			return ui.VStack(
				form.Auto(form.AutoOptions{Window: wnd}, thingState),
				ui.PrimaryButton(func() {
					fmt.Printf("Thing: %v\n", thingState.Get())
				}).Title("print"),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())
		}))

	}).Run()
}
