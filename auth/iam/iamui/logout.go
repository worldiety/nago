package iamui

import (
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func Logout(wnd core.Window, service *iam.Service) core.Component {
	service.Logout(wnd.SessionID())
	wnd.UpdateSubject(service.Subject(wnd.SessionID()))

	return ui.NewFlexContainer(func(flexContainer *ui.FlexContainer) {
		flexContainer.Elements().Append(
			ui.NewVBox(func(vbox *ui.VBox) {
				vbox.Append(
					ui.MakeText("Sie wurden sicher abgemeldet."),
					ui.NewActionButton("zur Startseite", func() {
						wnd.Navigation().ResetTo(".", nil)
					}),
				)
			}),
		)
	})
}
