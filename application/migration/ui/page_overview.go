// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uimigration

import (
	"github.com/worldiety/i18n"
	"github.com/worldiety/i18n/date"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/migration"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"golang.org/x/text/language"
)

var (
	TextMigrations = i18n.MustString(
		"nago.migration.text",
		i18n.Values{
			language.English: "Inspect the status of internal technical data migrations.",
			language.German:  "Prüfen Sie den Status von internen technischen Datenmigrationen.",
		},
	)
)

func PageOverview(wnd core.Window, migs *migration.Migrations) core.View {
	if !wnd.Subject().HasPermission(PermViewMigration) {
		return alert.BannerError(user.PermissionDeniedError(PermViewMigration))
	}

	return ui.VStack(
		ui.H1(rstring.LabelMigrations.Get(wnd)),
		dataview.FromData(wnd, dataview.Data[migration.Status, migration.Version]{
			FindAll:  migs.Versions(),
			FindByID: migs.Status,
			Fields: []dataview.Field[migration.Status]{
				{
					ID:   "version",
					Name: rstring.LabelVersion.Get(wnd),
					Map: func(obj migration.Status) core.View {
						return ui.Text(string(obj.Version))
					},
				},
				{
					ID:   "status",
					Name: rstring.LabelState.Get(wnd),
					Map: func(obj migration.Status) core.View {
						if obj.Installed {
							return ui.Text(rstring.LabelInstalled.Get(wnd))
						}

						if obj.Error != "" {
							return ui.Text(rstring.LabelError.Get(wnd))
						}

						return ui.Text(rstring.LabelPending.Get(wnd))
					},
				},

				{
					ID:   "script",
					Name: rstring.LabelScript.Get(wnd),
					Map: func(obj migration.Status) core.View {
						return ui.Text(obj.Script)
					},
				},

				{
					ID:   "date",
					Name: rstring.LabelCreatedAt.Get(wnd),
					Map: func(obj migration.Status) core.View {
						return ui.Text(date.Format(wnd.Locale(), date.TimeMinute, obj.InstalledAt.Time(wnd.Location())))
					},
				},

				{
					ID:   "desc",
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj migration.Status) core.View {
						if obj.Error != "" {
							return ui.Text(obj.Error)
						}

						return ui.Text("")
					},
				},
			},
		}).SelectOptions(dataview.SelectOption[migration.Version]{
			Name: rstring.ActionExecute.Get(wnd),
			Action: func(selected []migration.Version) error {
				for _, version := range selected {
					if err := migs.ReApply(wnd.Context(), version); err != nil {
						return err
					}
				}

				return nil
			},
			Visible: func(selected []migration.Version) bool {
				return wnd.Subject().HasPermission(PermViewReApply)
			},
		}),
	).FullWidth().Alignment(ui.Leading)
}
