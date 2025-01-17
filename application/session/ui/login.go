package uisession

import (
	"errors"
	"fmt"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

type SendPasswordResetMail func(email user.Email) error
type SendVerificationMail func(uid user.ID) error

func Login(
	wnd core.Window,
	loginFn session.Login,
	su user.SysUser,
	findByMail user.FindByMail,
	sendResetPwdMail SendPasswordResetMail,
	sendVerifyMail SendVerificationMail,
) core.View {
	if wnd.Subject().Valid() {
		return alert.Banner("Login", "Sie sind bereits eingeloggt.")
	}

	login := core.AutoState[string](wnd)
	password := core.AutoState[string](wnd)
	infoText := core.AutoState[string](wnd)
	presentMailInfo := core.AutoState[bool](wnd)
	invalidLoginText := core.AutoState[string](wnd)
	verificationDialogPresented := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H1("Login"),
		alert.Dialog(
			"Login nicht möglich",
			ui.Text("Das Konto muss zuerst bestätigt werden."),
			verificationDialogPresented,

			alert.Custom(
				func(close func(closeDlg bool)) core.View {
					return ui.SecondaryButton(func() {
						close(true)
					}).Title("Abbrechen")
				},
			),

			alert.Custom(
				func(close func(closeDlg bool)) core.View {
					return ui.PrimaryButton(func() {
						optUsr, err := findByMail(su(), user.Email(login.Get()))
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						if optUsr.IsNone() {
							// security note: don't expose knowledge whether the user exists or not
							return
						}

						if err := sendVerifyMail(optUsr.Unwrap().ID); err != nil {
							alert.ShowBannerError(wnd, err)
						}
						close(true)
					}).Title("Verifikationslink anfragen")
				},
			),
		),
		alert.Dialog("Hinweis", ui.Text(infoText.Get()), presentMailInfo, alert.Ok()),
		ui.Text(invalidLoginText.String()).Color(ui.SE0),
		ui.TextField("E-Mail Adresse", login.Get()).InputValue(login),
		ui.PasswordField("Kennwort", password.Get()).InputValue(password),
		ui.TertiaryButton(func() {
			if !user.Email(login.Get()).Valid() {
				infoText.Set(fmt.Sprintf("Die E-Mail-Adresse '%s' hat ein nicht unterstütztes Format.", login.Get()))
				presentMailInfo.Set(true)
				return
			}

			if sendResetPwdMail != nil {
				if err := sendResetPwdMail(user.Email(login.Get())); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				infoText.Set(fmt.Sprintf("Ein E-Mail mit einem Link zum Zurücksetzen wurde an '%s' gesendet. Prüfen Sie ihr Postfach.", login.Get()))
			}

			presentMailInfo.Set(true)
		}).Title("Kennwort vergessen"),
		ui.PrimaryButton(func() {
			ok, err := loginFn(wnd.Session().ID(), user.Email(login.Get()), user.Password(password.Get()))
			if err != nil {
				if errors.Is(err, user.EMailNotVerifiedErr) {
					verificationDialogPresented.Set(true)
					return
				}

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
