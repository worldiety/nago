// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgmigration

import (
	"log/slog"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/migration"
	uimigration "go.wdy.de/nago/application/migration/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type Module struct {
	Migrations *migration.Migrations
	Pages      uimigration.Pages
}

// Enable just installs the migration overview page into the admin section and injects the module into the
// context. The [application.Configurator] always executes and applies migrations anyway and is not
// affected if this module is enabled or not.
func Enable(cfg *application.Configurator) (Module, error) {
	mod, ok := core.FromContext[Module](cfg.Context(), "")
	if ok {
		return mod, nil
	}

	mg, err := cfg.Migrations()
	if err != nil {
		return Module{}, err
	}

	mod.Migrations = mg
	mod.Pages = uimigration.Pages{
		Overview: "admin/migration/overview",
	}

	cfg.RootViewWithDecoration(mod.Pages.Overview, func(wnd core.Window) core.View {
		return uimigration.PageOverview(wnd, mg)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var grp admin.Group
		if !subject.HasPermission(uimigration.PermViewMigration) {
			return grp
		}

		grp.Title = rstring.LabelMigrations.Get(subject)
		grp.Entries = append(grp.Entries, admin.Card{
			Title:  rstring.LabelOverview.Get(subject),
			Text:   uimigration.TextMigrations.Get(subject),
			Target: mod.Pages.Overview,
		})

		return grp
	})

	cfg.AddContextValue(core.ContextValue("nago.migrations", mod))
	slog.Info("installed migration module")

	return mod, nil
}
