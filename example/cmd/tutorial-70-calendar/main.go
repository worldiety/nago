// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/calendar"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_70")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("hello world"),
				calendar.Calendar(
					calendar.Event{
						From: calendar.Instant{
							At: time.Date(2025, 07, 11, 0, 0, 0, 0, time.Local),
						},
						To: calendar.Instant{
							At: time.Date(2025, 8, 11, 0, 0, 0, 0, time.Local),
						},
						Label: "Some Event",
						Lane: calendar.Lane{
							Label: "Torben",
						},
					},

					calendar.Event{
						From: calendar.Instant{
							At: time.Date(2025, 2, 1, 0, 0, 0, 0, time.Local),
						},
						To: calendar.Instant{
							At: time.Date(2025, 8, 31, 0, 0, 0, 0, time.Local),
						},
						Label: "Some other Event",
						Lane: calendar.Lane{
							Label: "Torben",
						},
					},
				).ViewPort(calendar.Year(2025)),
			).FullWidth()

		})
	}).Run()
}
