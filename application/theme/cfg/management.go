// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package cfgtheme installs a page in the admin center, on which a user chooses the colors of the theme. The
// colors which an application proposes are set with [application.Configurator.SetThemeDefaults] and apply until a
// user saves own colors.
package cfgtheme

import (
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/theme"
	uitheme "go.wdy.de/nago/application/theme/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

var (
	StrAdminGroup    = i18n.MustString("nago.theme.admin.group", i18n.Values{language.English: "Appearance", language.German: "Darstellung"})
	StrAdminCardDesc = i18n.MustString("nago.theme.admin.colors_desc", i18n.Values{language.English: "Adjust the colors of the light and the dark mode.", language.German: "Farben des Light- und Dark-Modes anpassen."})
)

type Management struct {
	UseCases theme.UseCases
	Pages    uitheme.Pages
}

// Enable installs the theme module. It is safe to call it multiple times.
func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	themes, err := cfg.ThemeManagement()
	if err != nil {
		return Management{}, err
	}

	management = Management{
		UseCases: themes.UseCases,
		Pages: uitheme.Pages{
			Colors: "admin/theme/colors",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.Colors, func(wnd core.Window) core.View {
		return uitheme.PageColors(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		return admin.Group{
			Title: StrAdminGroup.Get(subject),
			Entries: []admin.Card{
				{
					Title:      uitheme.StrColorsTitle.Get(subject),
					Text:       StrAdminCardDesc.Get(subject),
					Target:     management.Pages.Colors,
					Permission: theme.PermUpdateColors,
				},
			},
		}
	})

	cfg.AddContextValue(core.ContextValue("nago.theme.management", management))

	slog.Info("installed theme module")
	return management, nil
}
