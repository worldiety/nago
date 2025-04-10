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

		c.AddSystemService("nago.settings.global.load", uc.LoadGlobal)

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
