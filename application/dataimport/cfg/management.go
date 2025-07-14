// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgdataimport

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/dataimport"
	uidataimport "go.wdy.de/nago/application/dataimport/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
)

type Management struct {
	Pages    uidataimport.Pages
	UseCases dataimport.UseCases
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	management.Pages = uidataimport.Pages{
		PageStaging:      "admin/data/staging",
		PageStagings:     "admin/data/stagings",
		PageSelectParser: "admin/data/select-parser",
		PageEntry:        "admin/data/entry",
	}

	stagingStore, err := cfg.EntityStore("nago.dataimport.staging")
	if err != nil {
		return management, fmt.Errorf("cannot create staging store: %w", err)
	}

	stagingRepo := json.NewSloppyJSONRepository[dataimport.Staging](stagingStore)

	entryStore, err := cfg.EntityStore("nago.dataimport.entry")
	if err != nil {
		return management, fmt.Errorf("cannot create entry store: %w", err)
	}

	entryRepo := json.NewSloppyJSONRepository[dataimport.Entry](entryStore)

	management.UseCases = dataimport.NewUseCases(stagingRepo, entryRepo)
	cfg.AddContextValue(core.ContextValue("nago.dataimport", management))

	cfg.RootViewWithDecoration(management.Pages.PageStagings, func(wnd core.Window) core.View {
		return uidataimport.PageStagings(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.PageSelectParser, func(wnd core.Window) core.View {
		return uidataimport.PageSelectParser(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.PageStaging, func(wnd core.Window) core.View {
		return uidataimport.PageStaging(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.PageEntry, func(wnd core.Window) core.View {
		return uidataimport.PageEntry(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var grp admin.Group
		if !subject.HasPermission(dataimport.PermFindImporters) {
			return grp
		}

		grp.Title = "Daten Importe"
		for imp, err := range management.UseCases.FindImporters(subject) {
			if err != nil {
				slog.Error("failed to find all importers", "subject", subject, "err", err.Error())
				continue
			}

			desc := imp.Configuration()

			grp.Entries = append(grp.Entries, admin.Card{
				Title:        desc.Name,
				Text:         desc.Description,
				Target:       management.Pages.PageStagings,
				TargetParams: core.Values{"importer": string(imp.Identity())},
				Permission:   dataimport.PermFindImporters,
			})
		}

		return grp
	})

	slog.Info("installed data import management")

	return management, nil
}
