// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgcms

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/cms"
	uicms "go.wdy.de/nago/application/cms/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
)

type Management struct {
	UseCases cms.UseCases
	Pages    uicms.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := application.SystemServiceFor[Management](cfg, "")
	if ok {
		return management, nil
	}

	docStore, err := cfg.EntityStore("nago.cms.document")
	if err != nil {
		return Management{}, err
	}

	repo := json.NewSloppyJSONRepository[cms.PDocument](docStore)
	uc, err := cms.NewUseCases(repo)
	if err != nil {
		return Management{}, err
	}

	management = Management{
		UseCases: uc,
		Pages: uicms.Pages{
			Editor: "admin/cmd/editor",
		},
	}

	cfg.RootView(management.Pages.Editor, func(wnd core.Window) core.View {
		return uicms.PageEditor(wnd, management.UseCases)
	})

	prefixPageSlug := core.NavigationPath("page")
	cfg.RootViewWithDecoration(prefixPageSlug+"/*", func(wnd core.Window) core.View {
		return uicms.RenderPage(wnd, prefixPageSlug, management.UseCases.FindBySlug)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {

		return admin.Group{
			Title: "Statische Seiten",
			Entries: []admin.Card{
				{
					Title:      "CMS",
					Text:       "Verwaltung von statischen Seiten mit dem Content Management System.",
					Target:     management.Pages.Editor,
					Permission: cms.PermFindAll,
				},
			},
		}
	})

	cfg.AddSystemService("nago.cms", management)

	slog.Info("installed cms management")

	return management, nil
}
