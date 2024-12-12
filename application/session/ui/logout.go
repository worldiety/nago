package uisession

import (
	"fmt"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func Logout(wnd core.Window, logoutFn session.Logout) core.View {
	return ui.VStack(
		ui.IfElse(wnd.Subject().Valid(),
			ui.VStack(
				ui.Text(fmt.Sprintf("Sie sind derzeit als %s angemeldet.", wnd.Subject().Name())).TextAlignment(ui.TextAlignCenter),
				ui.PrimaryButton(func() {
					_, err := logoutFn(session.ID(wnd.SessionID()))
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					wnd.UpdateSubject(auth.InvalidSubject{})
				}).Title("Jetzt abmelden"),
			).Gap(ui.L16),
			ui.Text("Sie sind abgemeldet.").TextAlignment(ui.TextAlignCenter),
		),
	).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
}
