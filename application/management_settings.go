// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/settings"
	uisettings "go.wdy.de/nago/application/settings/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
)

// SettingsManagement is a nago system(Settings Management).
// It provides centralized configuration for global and per-user settings and has two main responsibilities:
//
//  1. Manage general application-level user settings such as:
//     - free registration (enable/disable)
//     - forgot password functionality
//     - domain whitelist for registration
//     - default roles and groups for new and anonymous users
//
//  2. Provide an extension mechanism for other systems to define and persist
//     their own settings. Any type implementing settings.GlobalSettings or settings.UserSettings
//     is automatically integrated into the Admin Center UI with generated forms.
//
// Examples:
//   - User Management: GDPR consent texts
//   - Theme Management: global theme configuration
//   - Schedule Management: job lifetime and cleanup rules
//
// Settings are stored persistently and can be loaded and written via UseCases:
//
//	settings := settings.ReadGlobal[user.Settings](cfg.SettingsManagement().UseCases.LoadGlobal)
//	option.MustZero(cfg.SettingsManagement().UseCases.StoreGlobal(user.SU(), settings))
type SettingsManagement struct {
	UseCases settings.UseCases
	Pages    uisettings.Pages
}

func (c *Configurator) SettingsManagement() (SettingsManagement, error) {
	if c.settingsManagement == nil {
		globalStore, err := c.EntityStore("nago.settings.global")
		if err != nil {
			return SettingsManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		globalRepo := json.NewSloppyJSONRepository[settings.StoreBox[settings.GlobalSettings]](globalStore)

		usrStore, err := c.EntityStore("nago.settings.user")
		if err != nil {
			return SettingsManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		usrRepo := json.NewSloppyJSONRepository[settings.StoreBox[settings.UserSettings]](usrStore)

		uc := settings.NewUseCases(globalRepo, usrRepo, c.EventBus())
		c.settingsManagement = &SettingsManagement{
			UseCases: uc,
			Pages: uisettings.Pages{
				PageSettings: "admin/settings/global",
			},
		}

		c.AddContextValue(core.ContextValue("nago.settings.global.load", uc.LoadGlobal))

		c.RootViewWithDecoration(c.settingsManagement.Pages.PageSettings, func(wnd core.Window) core.View {
			return uisettings.PageSettings(wnd, uc.LoadGlobal, uc.StoreGlobal)
		})

		c.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
			if err := subject.Audit(settings.PermLoadGlobal); err != nil {
				return admin.Group{}
			}

			decl, ok := enum.DeclarationFor[settings.GlobalSettings]()
			if !ok {
				// by default, the package has never registered any type
				return admin.Group{}
			}

			grp := admin.Group{
				Title: "Einstellungen",
			}
			for variant := range decl.Variants() {
				meta := settings.ReadMetaData(variant)
				grp.Entries = append(grp.Entries, admin.Card{
					Title:        meta.Title,
					Text:         meta.Description,
					Target:       c.settingsManagement.Pages.PageSettings,
					TargetParams: core.Values{"type": settings.TypeIdent(variant)},
					Permission:   settings.PermLoadGlobal,
				})
			}

			return grp
		})
	}

	return *c.settingsManagement, nil
}
