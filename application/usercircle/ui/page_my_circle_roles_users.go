package uiusercircles

import (
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
	"os"
	"slices"
)

func PageMyCircleRolesUsers(
	wnd core.Window,
	pages Pages,
	useCases usercircle.UseCases,
	findRoleById role.FindByID,
) core.View {
	optRole, err := findRoleById(user.SU(), role.ID(wnd.Values()["role"])) // security note: by definition, we are allowed to see
	if err != nil {
		return alert.BannerError(err)
	}

	if optRole.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	myRole := optRole.Unwrap()

	circle, err := loadMyCircle(wnd, useCases)
	if err != nil {
		return alert.BannerError(err)
	}

	return viewUsers(wnd, "Rolle / "+myRole.Name, useCases, func(usr user.User) bool {
		return slices.Contains(usr.Roles, myRole.ID)
	}, func(users []user.User) {
		for _, usr := range users {
			if err := useCases.MyCircleRolesAdd(wnd.Subject(), circle.ID, usr.ID, myRole.ID); err != nil {
				alert.ShowBannerError(wnd, err)
			}
		}
	},
		func(users []user.User) {
			for _, usr := range users {
				if err := useCases.MyCircleRolesRemove(wnd.Subject(), circle.ID, usr.ID, myRole.ID); err != nil {
					alert.ShowBannerError(wnd, err)
				}
			}
		},
	)
}
