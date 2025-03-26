// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiusercircles

import (
	"fmt"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"os"
	"slices"
)

func PageMyCircleLicensesUsers(
	wnd core.Window,
	pages Pages,
	useCases usercircle.UseCases,
	findLicByID license.FindUserLicenseByID,
	assignLicense user.AssignUserLicense,
	unassignLicense user.UnassignUserLicense,
	countLicense user.CountAssignedUserLicense,
) core.View {
	optLic, err := findLicByID(user.SU(), license.ID(wnd.Values()["license"])) // security note: by definition, we are allowed to see
	if err != nil {
		return alert.BannerError(err)
	}

	if optLic.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	myLicense := optLic.Unwrap()

	assigned, _ := countLicense(user.SU(), myLicense.ID)

	return ui.VStack(viewUsers(wnd, "Lizenz / "+myLicense.Name, useCases, func(usr user.User) bool {
		return slices.Contains(usr.Licenses, myLicense.ID)
	}, func(users []user.User) {
		for _, usr := range users {
			// security note: by circle definition, we are allowed to assign and we are protected by viewUsers
			if _, err := assignLicense(user.SU(), usr.ID, myLicense.ID); err != nil {
				alert.ShowBannerError(wnd, err)
			}
		}
	},
		func(users []user.User) {
			for _, usr := range users {
				// security note: by circle definition, we are allowed to unassign and we are protected by viewUsers
				if err := unassignLicense(user.SU(), usr.ID, myLicense.ID); err != nil {
					alert.ShowBannerError(wnd, err)
				}
			}
		},
	),
		ui.Text(fmt.Sprintf("Insgesamt %d von %d %s Lizenzen verwendet.", assigned, myLicense.MaxUsers, myLicense.Name)),
	).Gap(ui.L8).Alignment(ui.Leading).FullWidth()
}
