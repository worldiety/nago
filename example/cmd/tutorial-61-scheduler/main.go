// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	"fmt"
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
		cfg.SetApplicationID("de.worldiety.tutorial_61")

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

		var n int
		option.MustZero(schedulers.UseCases.Configure(user.SU(), scheduler.Options{
			ID:          "test.repeat",
			Name:        "Wiederholter Job",
			Description: "Schreibt viele Log-Einträge und schlägt jeden dritten Lauf fehl.",
			Kind:        scheduler.Schedule,
			Defaults: scheduler.Settings{
				PauseTime: 20 * time.Second,
			},
			Runner: func(ctx context.Context) error {
				n++
				log := logging.FromContext(ctx)
				for i := range 250 {
					log.Info("processing item", "item", i)
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(100 * time.Millisecond):
					}
				}
				log.WithGroup("smtp").Warn("slow response", "host", "mail.example.org")
				if n%3 == 0 {
					return fmt.Errorf("smtp: dial tcp 10.0.3.12:587: i/o timeout")
				}
				return nil
			},
			Actions: []scheduler.CustomAction{{Title: "Outbox leeren", Action: func(ctx context.Context) {}}},
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
