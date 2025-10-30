// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrClearCacheBtn     = i18n.MustString("nago.ai.admin.maintenance.action_clear", i18n.Values{language.English: "Clear cache", language.German: "Cache löschen"})
	StrClearCacheText    = i18n.MustString("nago.ai.admin.maintenance.desc_clear", i18n.Values{language.English: "Deletes all cache data for all configured providers and rebuild it from providers. All data which is not available remote, like user ownerships are also removed and cannot be restored.", language.German: "Löscht alle Cache-Daten für alle konfigurierten Anbieter und baut sie aus den Anbietern neu auf. Alle Daten, die nicht remote verfügbar sind, wie z. B. Benutzerrechte, werden ebenfalls entfernt und können nicht wiederhergestellt werden."})
	StrClearCacheConfirm = i18n.MustString("nago.ai.admin.maintenance.action_clear_confirm", i18n.Values{language.English: "Do you really want to delete the cache and all related data? This will cause a data loss.", language.German: "Möchten Sie den Cache und alle zugehörigen Daten wirklich löschen? Dies führt zu einem Datenverlust."})

	StrReloadProviderBtn  = i18n.MustString("nago.ai.admin.maintenance.action_reload", i18n.Values{language.English: "Reload provider", language.German: "Provider initialisieren"})
	StrReloadProviderText = i18n.MustString("nago.ai.admin.maintenance.desc_reload", i18n.Values{language.English: "Reloads all providers configurations and eventually re-creates caches based on the actual local and remote state.", language.German: "Lädt alle Anbieterkonfigurationen neu und erstellt schließlich Caches basierend auf dem aktuellen lokalen und Remote-Status neu."})
)

func PageMaintenance(wnd core.Window, uc ai.UseCases) core.View {
	deletePresented := core.AutoState[bool](wnd)
	syncActive := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H1(rstring.LabelMaintenance.Get(wnd)),
		alert.Dialog(StrClearCacheBtn.Get(wnd), ui.Text(StrClearCacheConfirm.Get(wnd)), deletePresented, alert.Cancel(nil), alert.Confirm(func() (close bool) {
			if err := uc.ClearCache(wnd.Subject()); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			return true
		})),
		ui.VStack(
			ui.HStack(
				ui.Text(StrClearCacheText.Get(wnd)),
			).FullWidth().Alignment(ui.Leading),
			ui.PrimaryButton(func() {
				deletePresented.Set(true)
			}).Title(StrClearCacheBtn.Get(wnd)),
		).FullWidth().Alignment(ui.Trailing).Gap(ui.L4),

		ui.HLine(),
		ui.VStack(
			ui.HStack(
				ui.Text(StrReloadProviderText.Get(wnd)),
			).FullWidth().Alignment(ui.Leading),
			ui.PrimaryButton(func() {
				syncActive.Set(true)
				xsync.Go(func() error {
					return uc.ReloadProvider(wnd.Subject(), ai.ReloadProviderOptions{LoadAll: true})
				}, func(err error) {
					wnd.Post(func() {
						syncActive.Set(false)
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}
					})
				})

			}).Title(StrReloadProviderBtn.Get(wnd)).Enabled(!syncActive.Get()),
		).FullWidth().Alignment(ui.Trailing).Gap(ui.L4),
	).
		Alignment(ui.Leading).
		Frame(ui.Frame{}.Larger())
}
