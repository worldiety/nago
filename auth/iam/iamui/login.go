package iamui

import (
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
)

func Login(wnd core.Window, modals ui.ModalOwner, service *iam.Service) core.Component {
	return ui.NewFlexContainer(func(flexContainer *ui.FlexContainer) {
		flexContainer.Elements().Append(
			ui.NewVBox(func(vbox *ui.VBox) {
				var mailLogin *ui.TextField
				var pwdLogin *ui.PasswordField
				var errMsg *ui.Text
				vbox.Append(
					ui.NewText(func(text *ui.Text) {
						errMsg = text
						text.Color().Set("red")
					}),
					ui.NewTextField(func(tf *ui.TextField) {
						mailLogin = tf
						if email, ok := service.Subject(wnd.SessionID()).(interface{ EMail() iam.Email }); ok {
							tf.Value().Set(string(email.EMail()))
						}
						tf.Label().Set("E-Mail Adresse")
					}),
					ui.NewPasswordField(func(passwordField *ui.PasswordField) {
						pwdLogin = passwordField
						passwordField.Label().Set("Kennwort")
					}),
					ui.NewText(func(text *ui.Text) {
						text.Value().Set("Passwort vergessen")
						text.OnClick().Set(func() {
							xdialog.ShowMessage(modals, "Die Self-Service Funktion steht nicht zur Verfügung. Bitte wenden Sie sich an Ihren Administrator.")
						})
					}),
					ui.NewButton(func(btn *ui.Button) {
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
	})

}
