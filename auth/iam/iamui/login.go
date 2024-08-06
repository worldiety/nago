package iamui

import (
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/uilegacy"
	"go.wdy.de/nago/presentation/uix/xdialog"
)

func Login(wnd core.Window, modals uilegacy.ModalOwner, service *iam.Service) core.View {
	return uilegacy.NewFlexContainer(func(flexContainer *uilegacy.FlexContainer) {
		flexContainer.ElementSize().Set(ora.ElementSizeLarge)
		flexContainer.Children().Append(
			uilegacy.NewVStack(func(vbox *uilegacy.VStack) {
				var mailLogin *uilegacy.TextField
				var pwdLogin *uilegacy.PasswordField
				var errMsg *uilegacy.Text
				vbox.Append(
					uilegacy.NewText(func(text *uilegacy.Text) {
						errMsg = text
						text.SetColor(ora.Error)
					}),
					uilegacy.NewTextField(func(tf *uilegacy.TextField) {
						mailLogin = tf
						if email, ok := service.Subject(wnd.SessionID()).(interface{ EMail() iam.Email }); ok {
							tf.Value().Set(string(email.EMail()))
						}
						tf.Label().Set("E-Mail Adresse")
					}),
					uilegacy.NewPasswordField(func(passwordField *uilegacy.PasswordField) {
						pwdLogin = passwordField
						passwordField.Label().Set("Kennwort")
					}),
					uilegacy.NewButton(func(text *uilegacy.Button) {
						text.Caption().Set("Passwort vergessen")
						text.Style().Set(ora.Tertiary)
						text.Action().Set(func() {
							xdialog.ShowMessage(modals, "Die Self-Service Funktion steht nicht zur Verfügung. Bitte wenden Sie sich an Ihren Administrator.")
						})
					}),
					uilegacy.NewFlexContainer(func(flex *uilegacy.FlexContainer) {
						flex.Append(
							uilegacy.NewButton(func(btn *uilegacy.Button) {
								btn.Caption().Set("Anmelden")
								btn.Action().Set(func() {
									ok := service.Login(wnd.SessionID(), mailLogin.Value().Get(), pwdLogin.Value().Get())
									if !ok {
										errMsg.Value().Set("Der Benutzer existiert nicht, das Konto wurde deaktiviert oder das Kennwort ist falsch.")
									}

									if ok {
										errMsg.Value().Set("")
										wnd.Navigation().Back()
									}
								})
							}),
						)
					}),
				)
			}),
		)
	})

}
