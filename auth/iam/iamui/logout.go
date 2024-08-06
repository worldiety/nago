package iamui

import (
	"fmt"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/uilegacy"
)

func Logout(wnd core.Window, service *iam.Service) core.View {
	return uilegacy.NewFlexContainer(func(flexContainer *uilegacy.FlexContainer) {
		flexContainer.Children().Append(
			uilegacy.NewVStack(func(vstack *uilegacy.VStack) {
				if wnd.Subject().Valid() {
					//vstack.ItemsAlignment().Set(ora.ItemsCenter)
					//ui.VStackAlignCenter(vstack)
					vstack.Append(
						uilegacy.MakeText(fmt.Sprintf("Sie sind derzeit als %s angemeldet.", wnd.Subject().Name())),
						uilegacy.NewButton(func(btn *uilegacy.Button) {
							btn.Style().Set(ora.Primary)
							btn.Caption().Set("Jetzt abmelden")
							btn.Action().Set(func() {
								service.Logout(wnd.SessionID())
								wnd.UpdateSubject(service.Subject(wnd.SessionID()))
								vstack.Children = nil
								vstack.Append(
									uilegacy.MakeText("Sie wurden sicher abgemeldet."),
									uilegacy.NewActionButton("zur Startseite", func() {
										wnd.Navigation().ResetTo(".", nil)
									}),
								)
							})
						}))
				} else {
					vstack.Append(uilegacy.MakeText("Sie sind bereits abgemeldet."))
				}
			}),
		)
	})
}
