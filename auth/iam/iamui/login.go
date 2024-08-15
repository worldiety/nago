package iamui

import (
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func Login(wnd core.Window, service *iam.Service) core.View {
	login := core.AutoState[string](wnd)
	password := core.AutoState[string](wnd)
	noSelfServicePresented := core.AutoState[bool](wnd)
	invalidLoginText := core.AutoState[string](wnd)

	return ui.VStack(
		alert.Dialog("Hinweis", ui.Text("Die Self-Service Funktion steht nicht zur Verfügung. Bitte wenden Sie sich an Ihren Administrator."), noSelfServicePresented, alert.Ok()),
		ui.Text(invalidLoginText.String()).Color(ui.SE0),
		ui.TextField("E-Mail Adresse", "").InputValue(login),
		ui.PasswordField("Kennwort").InputValue(password),
		ui.TertiaryButton(func() {
			noSelfServicePresented.Set(true)
		}).Title("Kennwort vergessen"),
		ui.PrimaryButton(func() {
			ok := service.Login(wnd.SessionID(), login.Get(), password.Get())
			if !ok {
				invalidLoginText.Set("Der Benutzer existiert nicht, das Konto wurde deaktiviert oder das Kennwort ist falsch.")
			} else {
				wnd.Navigation().ForwardTo(".", nil)
			}
		}).Title("Anmelden"),
	).Gap(ui.L4).Frame(ui.Frame{}.MatchScreen())

}
