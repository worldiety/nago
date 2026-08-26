// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisession

import (
	"errors"
	"fmt"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/image"
	httpimage "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"golang.org/x/text/language"
)

var (
	StrSignInWithSSO = i18n.MustString("nago.iam.login.sign_in_with_sso", i18n.Values{language.English: "or sign in with", language.German: "oder anmelden mit"})
)

type SendPasswordResetMail func(email user.Email) error
type SendVerificationMail func(uid user.ID) error

func Login(
	wnd core.Window,
	loginFn session.Login,
	startNLSFlow session.StartNLSFlow,
	su user.SysUser,
	findByMail user.FindByMail,
	sendResetPwdMail SendPasswordResetMail,
	sendVerifyMail SendVerificationMail,
	loadGlobalSettings settings.LoadGlobal,
	registerPath core.NavigationPath,
) core.View {
	if wnd.Subject().Valid() {
		return ui.VStack(
			alert.Banner("Login", "Sie sind bereits eingeloggt.").Intent(alert.IntentOk),
			ui.PrimaryButton(func() {
				wnd.Navigation().ForwardTo(".", nil)
			}).Title("Zurück zur Hauptseite"),
		).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())
	}

	usrSettings := settings.ReadGlobal[user.Settings](loadGlobalSettings)
	themeSettings := settings.ReadGlobal[theme.Settings](loadGlobalSettings)

	var logoImg core.View
	if themeSettings.PageLogoDark != "" || themeSettings.PageLogoLight != "" {
		dark := httpimage.URI(themeSettings.PageLogoDark, image.FitCover, 512, 512)
		light := httpimage.URI(themeSettings.PageLogoLight, image.FitCover, 512, 512)
		logoImg = ui.Image().URIAdaptive(light, dark).ObjectFit(ui.FitContain).Frame(ui.Frame{Width: ui.Full, Height: ui.L64})
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

	hasAppIcon := themeSettings.AppIconLight != "" || themeSettings.AppIconDark != ""
	isMobile := wnd.Info().SizeClass < core.SizeClassMedium
	gridCols := 2
	if isMobile {
		gridCols = 1
	}

	content := ui.Grid(
		ui.GridCell(
			ui.VStack(
				ui.If(hasAppIcon,
					ui.Image().
						Adaptive(themeSettings.AppIconLight, themeSettings.AppIconDark).
						ObjectFit(ui.FitContain).
						Frame(ui.Frame{}.Size(ui.L48, ui.L48)),
				),
				ui.If(hasAppIcon, ui.Space(ui.L16)),
				ui.Text("Mit "+wnd.Application().Name()+"-Konto anmelden").Font(ui.HeadlineSmall).Hyphens(ui.HyphensAuto),
				ui.Text("Bitte Zugangsdaten eingeben"),
			).Alignment(ui.TopLeading),
		).Alignment(ui.Leading),

		ui.GridCell(
			ui.VStack(
				verificationDialog(wnd, verificationDialogPresented, su, findByMail, sendVerifyMail, login),
				logoImg,
				loginForm(login, emailErr, password, passwordErr, presentPasswordForgotten, triggerLoginAction),
				forgotPasswordLink(presentPasswordForgotten, usrSettings),

				ui.If(infoText.Get() != "" && presentPasswordForgotten.Get(), ui.Text(fmt.Sprintf("Ein E-Mail mit einem Link zum Zurücksetzen wurde an '%s' gesendet. Prüfen Sie ihr Postfach.", login.Get())).TextAlignment(ui.TextAlignCenter)),
				ui.LinkWithAction("zurück zur Anmeldung", func() {
					presentPasswordForgotten.Set(false)

				}).Font(ui.BodySmall).Visible(presentPasswordForgotten.Get()),

				ui.IfFunc(usrSettings.HasSSO(), func() core.View { return ssoLogin(wnd, startNLSFlow) }),
			).Gap(ui.L8).FullWidth().NoClip(true),
		),
	).
		Gap(ui.L40).
		Columns(gridCols).
		FullWidth().
		Heights("auto").
		Padding(ui.Padding{}.All(ui.L16))

	return ui.VStack( // we don't have a scaffold
		ui.VStack(
			ui.WindowTitle(rstring.ActionLogin.Get(wnd)),
			cardlayout.Card("").
				Body(
					ui.VStack(content).FullWidth(),
				).Footer(
				ui.HStack(
					submitForgotPasswordBtn(wnd, login, emailErr, presentPasswordForgotten, sendResetPwdMail, infoText),
					submitLoginBtn(wnd, triggerLoginAction, presentPasswordForgotten),
				),
			).
				Padding(ui.Padding{Top: ui.L40, Bottom: ui.L24}.Horizontal(ui.L40)).
				Frame(ui.Frame{MaxWidth: ui.L880}.FullWidth()),
			ui.TextLayout(
				ui.Text("Noch kein Konto? Hier gleich "),
				ui.LinkWithAction("registrieren!", func() {
					wnd.Navigation().ForwardTo(registerPath, nil)
				}),
			).Font(ui.BodySmall).Visible(usrSettings.SelfRegistration),
		).Gap(ui.L16),
	).Frame(ui.Frame{}.MatchScreen())
}

func verificationDialog(
	wnd core.Window,
	presented *core.State[bool],
	su user.SysUser,
	findByMail user.FindByMail,
	sendVerifyMail SendVerificationMail,
	login *core.State[string],
) core.View {
	return alert.Dialog(
		"Login nicht möglich",
		ui.Text("Das Konto muss zuerst bestätigt werden."),
		presented,

		alert.Custom(
			func(close func(closeDlg bool)) core.View {
				return ui.SecondaryButton(func() {
					close(true)
				}).Title(rstring.ActionCancel.Get(wnd))
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
	)
}

func loginForm(
	login *core.State[string],
	emailErr *core.State[string],
	password *core.State[string],
	passwordErr *core.State[string],
	presentPasswordForgotten *core.State[bool],
	triggerLoginAction func(),
) core.View {
	return ui.Form(
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
		).Gap(ui.L12).FullWidth(),
	).
		Autocomplete(true).
		ID("nago-form-login").
		Frame(ui.Frame{}.FullWidth())
}

func forgotPasswordLink(presented *core.State[bool], usrSettings user.Settings) core.View {
	return ui.LinkWithAction("Passwort vergessen", func() {
		presented.Set(true)

	}).Font(ui.BodySmall).Visible(usrSettings.SelfPasswordReset && !presented.Get())
}

func ssoLogin(wnd core.Window, startNLSFlow session.StartNLSFlow) core.View {
	return ui.VStack(
		ui.HStack(
			ui.HLine().Frame(ui.Frame{Width: ui.L64}),
			ui.Text(StrSignInWithSSO.Get(wnd)),
			ui.HLine().Frame(ui.Frame{Width: ui.L64}),
		).FullWidth().Gap(ui.L8).Padding(ui.Padding{}.Vertical(ui.L8)),

		ui.SecondaryButton(func() {
			uri, err := startNLSFlow(wnd.Session().ID())
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			core.HTTPOpen(wnd.Navigation(), core.URI(uri), "_self")
		}).Title("SSO").Frame(ui.Frame{}.FullWidth()),
	).FullWidth().Gap(ui.L8)
}

func submitForgotPasswordBtn(wnd core.Window, login *core.State[string], emailErr *core.State[string], presentPasswordForgotten *core.State[bool], sendResetPwdMail func(user.Email) error, infoText *core.State[string]) core.View {
	return ui.PrimaryButton(func() {
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
	}).Visible(presentPasswordForgotten.Get()).Title("Link per E-Mail senden")
}

func submitLoginBtn(wnd core.Window, loginAction func(), presentPasswordForgotten *core.State[bool]) core.View {
	return ui.PrimaryButton(loginAction).Visible(!presentPasswordForgotten.Get()).Title(rstring.ActionLogin.Get(wnd)).ID("nago-action-login")
}
