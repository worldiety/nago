// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiusercircles

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
	"os"
	"slices"
)

func PageMyCircleGroupsUsers(
	wnd core.Window,
	pages Pages,
	useCases usercircle.UseCases,
	findGroupById group.FindByID,
) core.View {
	optGroup, err := findGroupById(user.SU(), group.ID(wnd.Values()["group"])) // security note: by definition, we are allowed to see
	if err != nil {
		return alert.BannerError(err)
	}

	if optGroup.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	myGroup := optGroup.Unwrap()

	circle, err := loadMyCircle(wnd, useCases)
	if err != nil {
		return alert.BannerError(err)
	}

	return viewUsers(wnd, "Grupppe / "+myGroup.Name, useCases, func(usr user.User) bool {
		return slices.Contains(usr.Groups, myGroup.ID)
	}, func(users []user.User) {
		for _, usr := range users {
			if err := useCases.MyCircleGroupsAdd(wnd.Subject(), circle.ID, usr.ID, myGroup.ID); err != nil {
				alert.ShowBannerError(wnd, err)
			}
		}
	},
		func(users []user.User) {
			for _, usr := range users {
				if err := useCases.MyCircleGroupsRemove(wnd.Subject(), circle.ID, usr.ID, myGroup.ID); err != nil {
					alert.ShowBannerError(wnd, err)
				}
			}
		},
	)
}
