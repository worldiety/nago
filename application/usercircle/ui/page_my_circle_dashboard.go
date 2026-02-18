// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiusercircles

import (
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

func PageMyCircleDashboard(wnd core.Window, pages Pages, useCases usercircle.UseCases) core.View {
	circle, err := loadMyCircle(wnd, useCases)
	if err != nil {
		return alert.BannerError(err)
	}

	return ui.VStack(
		ui.H1(circle.Name),
		cardlayout.Layout(

			cardlayout.Card("Benutzer").
				Body(ui.Text("Verwaltung der in diesem Kreis sichtbaren Nutzer.")).
				Footer(ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo(pages.MyCircleUsers, core.Values{"circle": string(circle.ID)})
				}).Title("Anzeigen")),

			ui.If(len(circle.Roles) > 0,
				cardlayout.Card("Rollen").
					Body(ui.Text("Verwaltung der Rollenzugehörigkeiten von Nutzer.")).
					Footer(ui.SecondaryButton(func() {
						wnd.Navigation().ForwardTo(pages.MyCircleRoles, core.Values{"circle": string(circle.ID)})
					}).Title("Anzeigen")),
			),

			ui.If(len(circle.Groups) > 0,
				cardlayout.Card("Gruppen").
					Body(ui.Text("Verwaltung der Gruppenzugehörigkeiten von Nutzern.")).
					Footer(ui.SecondaryButton(func() {
						wnd.Navigation().ForwardTo(pages.MyCircleGroups, core.Values{"circle": string(circle.ID)})
					}).Title("Anzeigen")),
			),
		),
	).Alignment(ui.Leading).FullWidth()
}
