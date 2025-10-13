// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfginspector

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/inspector"
	"go.wdy.de/nago/application/inspector/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
)

// Management is a Nago system(Inspector Management).
// It provides functionality to inspect and manage entity and blob stores.
// The system allows users to view and edit repository entries, download and delete blob files,
// and interact with stores through the Admin Center UI.
type Management struct {
	UseCases inspector.UseCases
	Pages    uiinspector.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	stores, err := cfg.Stores()
	if err != nil {
		return Management{}, err
	}

	management = Management{
		UseCases: inspector.NewUseCases(stores),
		Pages: uiinspector.Pages{
			PageDataInspector: "admin/inspector",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.PageDataInspector, func(wnd core.Window) core.View {
		return uiinspector.PageInspector(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		group := admin.Group{
			Title: "Inspektor",
			Entries: []admin.Card{
				{Title: "Stores", Text: "Stores bilden die Grundlage für Repositories. Es gibt spezialisierte Stores für Entities und Blobs.", Target: management.Pages.PageDataInspector, Permission: inspector.PermDataInspector},
			},
		}

		return group
	})
	cfg.AddContextValue(core.ContextValue("nago.inspector", management))

	slog.Info("installed inspector management")

	return management, nil
}
