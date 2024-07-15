package iamui

import (
	"fmt"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
)

func Logout(wnd core.Window, service *iam.Service) core.View {
	return ui.NewFlexContainer(func(flexContainer *ui.FlexContainer) {
		flexContainer.Children().Append(
			ui.NewVStack(func(vstack *ui.VStack) {
				if wnd.Subject().Valid() {
					//vstack.ItemsAlignment().Set(ora.ItemsCenter)
					//ui.VStackAlignCenter(vstack)
					vstack.Append(
						ui.MakeText(fmt.Sprintf("Sie sind derzeit als %s angemeldet.", wnd.Subject().Name())),
						ui.NewButton(func(btn *ui.Button) {
							btn.Style().Set(ora.Primary)
							btn.Caption().Set("Jetzt abmelden")
							btn.Action().Set(func() {
								service.Logout(wnd.SessionID())
								wnd.UpdateSubject(service.Subject(wnd.SessionID()))
								vstack.Children = nil
								vstack.Append(
									ui.MakeText("Sie wurden sicher abgemeldet."),
									ui.NewActionButton("zur Startseite", func() {
										wnd.Navigation().ResetTo(".", nil)
									}),
								)
							})
						}))
				} else {
					vstack.Append(ui.MakeText("Sie sind bereits abgemeldet."))
				}
			}),
		)
	})
}
