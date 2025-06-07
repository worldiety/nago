// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/scheduler"
	cfgscheduler "go.wdy.de/nago/application/scheduler/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {

	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_60")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		schedulers := option.Must(cfgscheduler.Enable(cfg))
		option.MustZero(schedulers.UseCases.Configure(user.SU(), scheduler.Options{
			Name:        "test.cron",
			Description: "cron job test",
			Kind:        scheduler.Cron,
			Defaults: scheduler.Settings{
				CronHour:   10,
				CronMinute: 40,
			},
			Runner: func(ctx context.Context) error {
				logging.FromContext(ctx).Info("Running cron")
				logging.FromContext(ctx).Info("some values", "key", "value")
				return nil
			},
			Actions: nil,
		}))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {

			return ui.VStack(
				ui.Text("scheduler demo, go to admin menu"),
			).
				Frame(ui.Frame{}.MatchScreen())

		})
	}).
		Run()
}
