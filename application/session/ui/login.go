package uisession

import (
	"errors"
	"fmt"
	"go.wdy.de/nago/application/image"
	httpimage "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
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
	loadGlobalSettings settings.LoadGlobal,
	registerPath core.NavigationPath,
) core.View {
	if wnd.Subject().Valid() {
		return alert.Banner("Login", "Sie sind bereits eingeloggt.")
	}

	usrSettings := settings.ReadGlobal[user.Settings](loadGlobalSettings)
	themeSettings := settings.ReadGlobal[theme.Settings](loadGlobalSettings)

	var logoImg core.View
	if themeSettings.PageLogoDark != "" || themeSettings.PageLogoLight != "" {
		dark := httpimage.URI(themeSettings.PageLogoDark, image.FitCover, 512, 512)
		light := httpimage.URI(themeSettings.PageLogoLight, image.FitCover, 512, 512)
		logoImg = ui.Image().URIAdaptive(light, dark).Frame(ui.Frame{Width: ui.Full, Height: ui.L64})
	}

	emailErr := core.AutoState[string](wnd)
	login := core.AutoState[string](wnd).Observe(func(newValue string) {
		if newValue == "" {
			emailErr.Set("")
			return
		}

		if !user.Email(newValue).Valid() {
			emailErr.Set("Diese E-Mail-Adresse ist ungültig.")
			return
		} else {
			emailErr.Set("")
		}
	})
	passwordErr := core.AutoState[string](wnd)
	password := core.AutoState[string](wnd)
	presentPasswordForgotten := core.AutoState[bool](wnd)
	verificationDialogPresented := core.AutoState[bool](wnd)
	infoText := core.AutoState[string](wnd)

	triggerLoginAction := func() {
		if !user.Email(login.Get()).Valid() {
			emailErr.Set("Diese E-Mail-Adresse ist ungültig.")
			return
		}

		if password.Get() == "" {
			passwordErr.Set("")
			return
		}

		ok, err := loginFn(wnd.Session().ID(), user.Email(login.Get()), user.Password(password.Get()))
		if err != nil {
			if errors.Is(err, user.EMailNotVerifiedErr) {
				verificationDialogPresented.Set(true)
				return
			}

			passwordErr.Set("Der Benutzer existiert nicht, das Konto wurde deaktiviert oder das Kennwort ist falsch.")
			return
		}
		if !ok {
			fmt.Println("cannot happen?")
		} else {
			password.Set("") // clean the password immediately from memory
			wnd.Navigation().ForwardTo(".", nil)
		}
	}

	return ui.VStack( // we don't have a scaffold
		ui.VStack(
			ui.WindowTitle("Anmelden"),
			cardlayout.Card("").
				Padding(ui.Padding{}.All(ui.L12)).
				Body(
					ui.VStack(
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
						logoImg,
						ui.Form(
							ui.VStack(
								ui.TextField("E-Mail Adresse", login.Get()).
									InputValue(login).
									ErrorText(emailErr.Get()).
									ID("nago-login").
									KeydownEnter(triggerLoginAction).
									Frame(ui.Frame{}.FullWidth()),

								ui.PasswordField("Kennwort", password.Get()).
									InputValue(password).
									ErrorText(passwordErr.Get()).
									ID("nago-password").
									KeydownEnter(triggerLoginAction).
									Frame(ui.Frame{}.FullWidth()).
									Visible(!presentPasswordForgotten.Get()),
							).Gap(ui.L4),
						).
							Autocomplete(true).
							ID("nago-form-login"),
						ui.LinkWithAction("Passwort vergessen", func() {
							presentPasswordForgotten.Set(true)

						}).Font(ui.Small).Visible(usrSettings.SelfPasswordReset && !presentPasswordForgotten.Get()),

						ui.If(infoText.Get() != "", ui.Text(fmt.Sprintf("Ein E-Mail mit einem Link zum Zurücksetzen wurde an '%s' gesendet. Prüfen Sie ihr Postfach.", login.Get())).TextAlignment(ui.TextAlignCenter)),
						ui.LinkWithAction("zurück zur Anmeldung", func() {
							presentPasswordForgotten.Set(false)

						}).Font(ui.Small).Visible(presentPasswordForgotten.Get()),
					).Gap(ui.L8),
				).Footer(
				ui.HStack(
					ui.PrimaryButton(func() {
						if !user.Email(login.Get()).Valid() {
							emailErr.Set("Diese E-Mail-Adresse ist ungültig.")
							return
						}

						if sendResetPwdMail != nil {
							if err := sendResetPwdMail(user.Email(login.Get())); err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}

						}

						infoText.Set(fmt.Sprintf("Eine E-Mail mit einem Link zum Zurücksetzen wurde an '%s' gesendet. Prüfen Sie ihr Postfach.", login.Get()))
					}).Visible(presentPasswordForgotten.Get()).Title("Link per E-Mail senden"),
					ui.PrimaryButton(triggerLoginAction).Visible(!presentPasswordForgotten.Get()).Title("Anmelden").ID("nago-action-login"),
				),
			),
			ui.TextLayout(
				ui.Text("Noch kein Konto? Hier gleich "),
				ui.LinkWithAction("registrieren!", func() {
					wnd.Navigation().ForwardTo(registerPath, nil)
				}),
			).Font(ui.Small).Visible(usrSettings.SelfRegistration),
		).Gap(ui.L16).Frame(ui.Frame{Width: ui.L320, Height: ""}), // "calc(100dvh - 7rem)"
	).Frame(ui.Frame{}.MatchScreen())
}
