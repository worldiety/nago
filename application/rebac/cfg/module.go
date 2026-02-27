// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgrebac

import (
	"log/slog"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/rebac"
	ucrebac "go.wdy.de/nago/application/rebac/uc"
	uirebac "go.wdy.de/nago/application/rebac/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type Module struct {
	Pages    uirebac.Pages
	DB       *rebac.DB
	UseCases ucrebac.UseCases
}

func Enable(cfg *application.Configurator) (Module, error) {
	mod, ok := core.FromContext[Module](cfg.Context(), "")
	if ok {
		return mod, nil
	}

	db, err := cfg.RDB()
	if err != nil {
		return Module{}, err
	}

	uc := ucrebac.NewUseCases(db)

	mod = Module{
		Pages: uirebac.Pages{
			Editor: "admin/rebac/editor",
		},
		DB:       db,
		UseCases: uc,
	}

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var grp admin.Group
		if !subject.HasPermission(ucrebac.PermFindAllResources) {
			return grp
		}

		grp.Title = uirebac.StrResourcesAndGrants.Get(subject)

		for resources, err := range uc.FindAllResources(subject) {
			if err != nil {
				slog.Error("failed to find all resources", "err", err.Error())
				continue
			}

			grp.Entries = append(grp.Entries, admin.Card{
				Title:        resources.Info(subject).Name,
				Text:         resources.Info(subject).Description,
				Target:       mod.Pages.Editor,
				TargetParams: core.Values{"resources": string(resources.Identity())},
			})
		}

		return grp
	})

	cfg.NoFooter(mod.Pages.Editor)
	
	cfg.RootViewWithDecoration(mod.Pages.Editor, func(wnd core.Window) core.View {
		return uirebac.PageEditor(wnd, uc)
	})

	cfg.AddContextValue(core.ContextValue("nago.rebac", mod))
	slog.Info("installed rebac module")

	return mod, nil
}
