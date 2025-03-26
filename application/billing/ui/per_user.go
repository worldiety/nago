// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uibilling

import (
	"fmt"
	"go.wdy.de/nago/application/billing"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/presentation/ui/tags"
	"slices"
)

func UserLicensePage(wnd core.Window, licenses billing.UserLicenses) core.View {
	usrStats, err := licenses(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}

	return ui.VStack(
		ui.H1("Kontingente Nutzer-Lizenzen"),
		list.List(
			ui.Each(slices.Values(usrStats.Stats), func(t billing.PerUserLicenseStats) core.View {
				var tmp []core.View

				tmp = append(tmp, ui.Text(fmt.Sprintf("%der Paket", t.License.MaxUsers)))
				switch {
				case t.Depleted():
					if t.License.Incentive != "" {
						tmp = append(tmp, ui.PrimaryButton(func() {
							wnd.Navigation().Open(core.URI(t.License.Incentive))
						}).Title("jetzt anfragen"))
					} else {
						tmp = append(tmp, tags.ColoredTextPill(ui.ColorSemanticWarn, "alle Lizenzen zugewiesen"))
					}
				case t.Overcommitted():
					tmp = append(tmp, tags.ColoredTextPill(ui.ColorSemanticError, "Lizensierung erforderlich"))
				default:
					tmp = append(tmp, tags.ColoredTextPill(ui.ColorSemanticGood, fmt.Sprintf("noch %d verfügbar", t.Avail())))
				}

				if t.License.Url != "" {
					tmp = append(tmp, ui.TertiaryButton(func() {
						core.HTTPOpen(wnd.Navigation(), core.URI(t.License.Url), "_blank")
					}).PreIcon(heroOutline.InformationCircle))
				}

				entry := list.Entry().
					Headline(t.License.Name).
					SupportingText(t.License.Description).
					Trailing(ui.HStack(tmp...).Gap(ui.L8))

				entry = entry.Leading(ui.ImageIcon(heroOutline.UserCircle))

				return entry
			})...,
		).Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).FullWidth()
}
