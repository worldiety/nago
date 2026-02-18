// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfgflow "go.wdy.de/nago/application/flow/cfg"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	"go.wdy.de/nago/application/migration"
	cfgmigration "go.wdy.de/nago/application/migration/cfg"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_88")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.Must(cfginspector.Enable(cfg))
		modMig := option.Must(cfgmigration.Enable(cfg)) // see, we have a migration system

		_, _ = cfg.RDB() //see, we have a rebac system, this is the database

		option.Must(cfgflow.Enable(cfg, cfgflow.Options{}))

		option.MustZero(modMig.Migrations.Declare(MyTestMigration{}, migration.Options{}))

		/*cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return
		})*/

	}).Run()
}

type MyTestMigration struct {
}

func (m MyTestMigration) Version() migration.Version {
	return migration.NewVersion(2026, time.January, 14, 16, 19, "TestMigration")
}

func (m MyTestMigration) Migrate(ctx context.Context) error {
	slog.Info("executed")
	return nil
}
