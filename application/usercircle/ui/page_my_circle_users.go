package uiusercircles

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

type userState struct {
	usr      user.User
	selected *core.State[bool]
	visible  bool
}

func PageMyCircleUsers(wnd core.Window, useCases usercircle.UseCases) core.View {
	_, err := loadMyCircle(wnd, useCases)
	if err != nil {
		return alert.BannerError(err)
	}

	return viewUsers(wnd, "Konten", useCases, func(usr user.User) bool {
		return true
	}, nil, nil)
}
