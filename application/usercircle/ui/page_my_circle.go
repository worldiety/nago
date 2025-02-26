package uiusercircles

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
	"os"
)

func PageMyCircle(wnd core.Window, useCases usercircle.UseCases) core.View {
	id := usercircle.ID(wnd.Values()["id"])
	optCircle, err := useCases.FindByID(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	if optCircle.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	circle := optCircle.Unwrap()

	return ui.VStack(
		ui.H1(circle.Name),
		list.List(ui.Each2(useCases.MyCircleMembers(wnd.Subject().ID(), circle.ID), func(usr user.User, err error) core.View {
			if err != nil {
				return alert.BannerError(err)
			}

			return list.Entry().
				Headline(usr.String()).
				SupportingText(string(usr.Email)).
				Leading(avatar.TextOrImage(usr.String(), usr.Contact.Avatar))
		})...).
			Caption(ui.Text("Nutzer in diesem Kreis")).
			Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).FullWidth()
}
