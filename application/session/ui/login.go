package uisession

import (
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func Login(wnd core.Window, loginFn session.Login) core.View {
	if wnd.Subject().Valid() {
		return alert.Banner("Login", "Sie sind bereits eingeloggt.")
	}

	login := core.AutoState[string](wnd)
	password := core.AutoState[string](wnd)
	noSelfServicePresented := core.AutoState[bool](wnd)
	invalidLoginText := core.AutoState[string](wnd)

	return ui.VStack(
		alert.Dialog("Hinweis", ui.Text("Die Self-Service Funktion steht nicht zur Verfügung. Bitte wenden Sie sich an Ihren Administrator."), noSelfServicePresented, alert.Ok()),
		ui.Text(invalidLoginText.String()).Color(ui.SE0),
		ui.TextField("E-Mail Adresse", login.Get()).InputValue(login),
		ui.PasswordField("Kennwort", password.Get()).InputValue(password),
		ui.TertiaryButton(func() {
			noSelfServicePresented.Set(true)
		}).Title("Kennwort vergessen"),
		ui.PrimaryButton(func() {
			ok, err := loginFn(session.ID(wnd.SessionID()), user.Email(login.Get()), user.Password(password.Get()))
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
			if !ok {
				invalidLoginText.Set("Der Benutzer existiert nicht, das Konto wurde deaktiviert oder das Kennwort ist falsch.")
			} else {
				password.Set("") // clean the password immediately from memory
				wnd.Navigation().ForwardTo(".", nil)
			}
		}).Title("Anmelden"),
	).Gap(ui.L4).Frame(ui.Frame{}.MatchScreen())

}
