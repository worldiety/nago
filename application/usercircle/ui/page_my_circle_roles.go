// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiusercircles

import (
	"fmt"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
)

func PageMyCircleRoles(wnd core.Window, pages Pages, useCases usercircle.UseCases, findRoleById role.FindByID) core.View {
	circle, err := loadMyCircle(wnd, useCases)
	if err != nil {
		return alert.BannerError(err)
	}

	var roles []role.Role
	for _, id := range circle.Roles {
		optRole, err := findRoleById(user.SU(), id) // security note: we are allowed by user circle definition
		if err != nil {
			return alert.BannerError(err)
		}

		if optRole.IsNone() {
			continue
		}

		roles = append(roles, optRole.Unwrap())
	}

	return ui.VStack(
		ui.H1(circle.Name+" / Rollen"),
		list.List(ui.ForEach(roles, func(t role.Role) core.View {
			return list.Entry().
				Headline(t.Name).
				SupportingText(t.Description + fmt.Sprintf(" (%d Berechtigungen)", len(t.Permissions))).
				Trailing(ui.ImageIcon(heroSolid.ChevronRight))
		})...).OnEntryClicked(func(idx int) {
			rle := roles[idx]
			wnd.Navigation().ForwardTo(pages.MyCircleRolesUsers, core.Values{"circle": string(circle.ID), "role": string(rle.ID)})
		}).
			Caption(ui.Text("In diesem Kreis sichtbare Rollen")).
			Footer(ui.Text(fmt.Sprintf("%d Rollen sind zur Verwaltung verfügbar", len(roles)))).
			Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).FullWidth()
}
