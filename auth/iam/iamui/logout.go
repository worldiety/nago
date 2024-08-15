package iamui

import (
	"fmt"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func Logout(wnd core.Window, service *iam.Service) core.View {
	return ui.VStack(
		ui.IfElse(wnd.Subject().Valid(),
			ui.VStack(
				ui.Text(fmt.Sprintf("Sie sind derzeit als %s angemeldet.", wnd.Subject().Name())),
				ui.PrimaryButton(func() {
					service.Logout(wnd.SessionID())
					wnd.UpdateSubject(service.Subject(wnd.SessionID()))
				}).Title("Jetzt abmelden"),
			),
			ui.Text("Sie sind abgemeldet."),
		),
	)
}
