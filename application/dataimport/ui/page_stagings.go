// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidataimport

import (
	"fmt"
	"os"
	"slices"

	"go.wdy.de/nago/application/dataimport"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/hero"
)

func PageStagings(wnd core.Window, ucImp dataimport.UseCases) core.View {
	id := importer.ID(wnd.Values()["importer"])
	optImp, err := ucImp.FindImporterByID(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	if optImp.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	imp := optImp.Unwrap()

	displayName, ok := core.FromContext[user.DisplayName](wnd.Context(), "")
	if !ok {
		displayName = func(uid user.ID) user.Compact {
			return user.Compact{ID: uid}
		}
	}

	stagings, err := xslices.Collect2(ucImp.FindStagingsForImporter(wnd.Subject(), imp.Identity()))
	if err != nil {
		return alert.BannerError(err)
	}

	slices.SortFunc(stagings, func(a, b dataimport.Staging) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	return ui.VStack(
		hero.Hero("Import Entwürfe - "+imp.Configuration().Name).
			Subtitle(imp.Configuration().Description).
			SideSVG(imp.Configuration().Image).
			Actions(ui.PrimaryButton(func() {
				if err := wnd.Subject().Audit(dataimport.PermImport); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
				wnd.Navigation().ForwardTo("admin/data/select-parser", core.Values{"importer": string(imp.Identity())})
			}).Title("Import Entwurf erstellen")),
		ui.Space(ui.L32),
		cardlayout.Layout(
			ui.ForEach(stagings, func(desc dataimport.Staging) core.View {
				name := "Import Entwurf"
				if desc.Name != "" {
					name = desc.Name
				}
				return cardlayout.Card(name).Body(
					ui.VStack(
						ui.Text(fmt.Sprintf("Erstellt am: %s", desc.CreatedAt.Format(xtime.GermanDateTime))),
						ui.Text(fmt.Sprintf("Erstellt von: %s", displayName(desc.CreatedBy).Displayname)),
						ui.Text(desc.Comment),
					).Alignment(ui.Leading),
				).Footer(ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo("admin/data/staging", core.Values{"stage": string(desc.ID)})
				}).Title("Öffnen"))
			})...,
		).Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).
		FullWidth()
}
