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
	"github.com/worldiety/enum"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/scheduler"
	cfgscheduler "go.wdy.de/nago/application/scheduler/cfg"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

var _ = enum.Variant[secret.Credentials, secret.BookStack]()

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "8fb8724f-e604-444c-9671-58d07dd76164"))

		option.MustZero(cfg.StandardSystems())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		schedulerManagement := std.Must(cfgscheduler.Enable(cfg))
		option.MustZero(schedulerManagement.UseCases.Configure(user.SU(), scheduler.Options{
			ID:          "my.test.scheduler",
			Name:        "One Shot hello world",
			Description: "This scheduler just runs once after start and is done",
			Kind:        scheduler.OneShot,
			Runner: func(ctx context.Context) error {
				log := scheduler.LoggerFrom(ctx)
				log.Info("hello world", "a", "b")
				return nil
			},
			Actions: []scheduler.CustomAction{
				{
					Title: "hello world",
					Action: func(ctx context.Context) {
						fmt.Println("hello world")
					},
				},
				{
					Title: "hello world",
					Action: func(ctx context.Context) {
						fmt.Println("hello world")
					},
				},
			},
		}))

		option.MustZero(schedulerManagement.UseCases.Configure(user.SU(), scheduler.Options{
			ID:          "my.test.failure",
			Name:        "Fail with random",
			Description: "This scheduler runs scheduled and fails in random ways",
			Kind:        scheduler.Schedule,
			Defaults: scheduler.Settings{
				StartDelay: time.Second,
				PauseTime:  time.Second * 10,
			},
			Runner: func(ctx context.Context) error {
				log := scheduler.LoggerFrom(ctx)
				log.Info("hello world")
				for range 100 {
					if ctx.Err() != nil {
						return ctx.Err()
					}

					time.Sleep(time.Millisecond * 500)

					r := time.Now().UnixMilli() % 1234
					switch r {
					case 0:
						panic("ops - started to panic")
					case 1:
						return fmt.Errorf("failed randomly")
					default:
						log.Info("the random source did not hit me", "value", r)
					}

				}
				return nil
			},
		}))

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			return ui.Text("hello service")

		}))

	}).Run()
}
