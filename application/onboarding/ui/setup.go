// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uionboarding

import (
	"errors"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/onboarding"
	uisession "go.wdy.de/nago/application/session/ui"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrTitle             = i18n.MustVarString("nago.onboarding.title_x", i18n.Values{language.German: "Willkommen bei {app}", language.English: "Welcome to {app}"})
	StrWelcome           = i18n.MustVarString("nago.onboarding.welcome_x", i18n.Values{language.German: "Diese Anwendung ist noch nicht eingerichtet. Zur Bestätigung senden wir einen Code an {email}.", language.English: "This application has not been set up yet. To confirm, we send a code to {email}."})
	StrStart             = i18n.MustString("nago.onboarding.start", i18n.Values{language.German: "Einrichtung starten", language.English: "Start setup"})
	StrCodeSubtitle      = i18n.MustVarString("nago.onboarding.code_subtitle_x", i18n.Values{language.German: "Wir haben einen Code an {email} gesendet.", language.English: "We have sent a code to {email}."})
	StrCode              = i18n.MustString("nago.onboarding.code", i18n.Values{language.German: "Code", language.English: "Code"})
	StrCodeHint          = i18n.MustVarString("nago.onboarding.code_hint_x", i18n.Values{language.German: "Gültig bis {time} Uhr.", language.English: "Valid until {time}."})
	StrConfirm           = i18n.MustString("nago.onboarding.confirm", i18n.Values{language.German: "Bestätigen", language.English: "Confirm"})
	StrResend            = i18n.MustString("nago.onboarding.resend", i18n.Values{language.German: "Code erneut senden", language.English: "Send code again"})
	StrProfileSubtitle   = i18n.MustString("nago.onboarding.profile_subtitle", i18n.Values{language.German: "Lege jetzt dein Konto an.", language.English: "Now create your account."})
	StrFirstname         = i18n.MustString("nago.onboarding.firstname", i18n.Values{language.German: "Vorname", language.English: "First name"})
	StrLastname          = i18n.MustString("nago.onboarding.lastname", i18n.Values{language.German: "Nachname", language.English: "Last name"})
	StrPassword          = i18n.MustString("nago.onboarding.password", i18n.Values{language.German: "Passwort", language.English: "Password"})
	StrPasswordRepeated  = i18n.MustString("nago.onboarding.password_repeated", i18n.Values{language.German: "Passwort wiederholen", language.English: "Repeat password"})
	StrPasswordsDiffer   = i18n.MustString("nago.onboarding.passwords_differ", i18n.Values{language.German: "Die Passwörter stimmen nicht überein.", language.English: "The passwords do not match."})
	StrRequired          = i18n.MustString("nago.onboarding.required", i18n.Values{language.German: "Bitte ausfüllen.", language.English: "Please fill in."})
	StrCreateAccount     = i18n.MustString("nago.onboarding.create_account", i18n.Values{language.German: "Konto anlegen", language.English: "Create account"})
	StrProblemTitle      = i18n.MustString("nago.onboarding.problem_title", i18n.Values{language.German: "Einrichtung nicht möglich", language.English: "Setup not possible"})
	StrProblem           = i18n.MustString("nago.onboarding.problem", i18n.Values{language.German: "Für diese Anwendung wurde kein Einrichtungsnutzer konfiguriert. Bitte wende dich an den Support.", language.English: "No setup user has been configured for this application. Please contact the support."})
	StrDoneTitle         = i18n.MustString("nago.onboarding.done_title", i18n.Values{language.German: "Einrichtung abgeschlossen", language.English: "Setup completed"})
	StrDone              = i18n.MustString("nago.onboarding.done", i18n.Values{language.German: "Diese Anwendung ist bereits eingerichtet.", language.English: "This application has already been set up."})
	StrContinue          = i18n.MustString("nago.onboarding.continue", i18n.Values{language.German: "Weiter", language.English: "Continue"})
	StrErrInvalidCode    = i18n.MustString("nago.onboarding.err_invalid_code", i18n.Values{language.German: "Der Code ist nicht gültig.", language.English: "The code is not valid."})
	StrErrCodeExpired    = i18n.MustString("nago.onboarding.err_code_expired", i18n.Values{language.German: "Der Code ist abgelaufen. Bitte fordere einen neuen an.", language.English: "The code has expired. Please request a new one."})
	StrErrTooManyAttempt = i18n.MustString("nago.onboarding.err_too_many_attempts", i18n.Values{language.German: "Zu viele falsche Eingaben. Bitte fordere einen neuen Code an.", language.English: "Too many wrong codes. Please request a new one."})
	StrErrNoCode         = i18n.MustString("nago.onboarding.err_no_code", i18n.Values{language.German: "Es ist kein Code gültig. Bitte fordere einen neuen an.", language.English: "No code is valid. Please request a new one."})
	StrErrCooldownX      = i18n.MustVarString("nago.onboarding.err_cooldown_x", i18n.Values{language.German: "Ein neuer Code kann ab {time} Uhr angefordert werden.", language.English: "A new code can be requested at {time}."})
	StrErrNotVerified    = i18n.MustString("nago.onboarding.err_not_verified", i18n.Values{language.German: "Die Bestätigung ist abgelaufen. Bitte fordere einen neuen Code an.", language.English: "The confirmation has expired. Please request a new code."})
)

// Setup is the default setup page.
func Setup(wnd core.Window, flow *Flow) core.View {
	themeSettings := core.GlobalSettings[theme.Settings](wnd)
	codeErr := core.AutoState[string](wnd)
	code := core.AutoState[string](wnd)

	firstname := core.AutoState[string](wnd)
	lastname := core.AutoState[string](wnd)
	password := core.AutoState[string](wnd)
	passwordRepeated := core.AutoState[string](wnd)
	errFirstname := core.AutoState[string](wnd)
	errLastname := core.AutoState[string](wnd)
	errPassword := core.AutoState[string](wnd)

	st := flow.State()
	app := wnd.Application().Name()
	title := StrTitle.Get(wnd, i18n.String("app", app))
	email := st.MaskedEmail()

	var subtitle string
	var body core.View
	var actions []core.View

	switch flow.Step() {
	case StepProblem:
		title = StrProblemTitle.Get(wnd)
		body = alert.Banner(StrProblemTitle.Get(wnd), StrProblem.Get(wnd)).Intent(alert.IntentError).Frame(ui.Frame{}.FullWidth())

	case StepDone:
		title = StrDoneTitle.Get(wnd)
		body = ui.Text(StrDone.Get(wnd))
		actions = append(actions, ui.PrimaryButton(func() {
			wnd.Navigation().ResetTo(".", nil)
		}).Title(StrContinue.Get(wnd)).Key("onboarding", "continue"))

	case StepWelcome:
		body = ui.Text(StrWelcome.Get(wnd, i18n.String("email", email)))
		actions = append(actions, ui.PrimaryButton(func() {
			codeErr.Set("")
			if err := flow.RequestCode(); err != nil {
				var cooldown onboarding.CooldownError
				if !errors.As(err, &cooldown) {
					alert.ShowBannerError(wnd, err)
					return
				}
				codeErr.Set(errorText(wnd, err))
			}
			code.Set("")
		}).Title(StrStart.Get(wnd)).Key("onboarding", "start"))

	case StepCode:
		subtitle = StrCodeSubtitle.Get(wnd, i18n.String("email", email))
		hint := ""
		if c, ok := flow.Code(); ok {
			hint = StrCodeHint.Get(wnd, i18n.String("time", c.ValidUntil.In(time.Local).Format("15:04")))
		}

		body = ui.TextField(StrCode.Get(wnd), code.Get()).
			InputValue(code).
			KeyboardType(ui.KeyboardPhone). // a numeric keypad, but unlike KeyboardInteger it keeps leading zeros
			SupportingText(hint).
			ErrorText(codeErr.Get()).
			ID("onboarding-code").
			FullWidth()

		actions = append(actions,
			ui.SecondaryButton(func() {
				codeErr.Set("")
				if err := flow.RequestCode(); err != nil {
					codeErr.Set(errorText(wnd, err))
				}
			}).Title(StrResend.Get(wnd)).Key("onboarding", "resend"),
			ui.PrimaryButton(func() {
				codeErr.Set("")
				if err := flow.VerifyCode(code.Get()); err != nil {
					codeErr.Set(errorText(wnd, err))
				}
			}).Title(StrConfirm.Get(wnd)).Key("onboarding", "confirm"),
		)

	case StepProfile:
		subtitle = StrProfileSubtitle.Get(wnd)
		body = ui.VStack(
			ui.TextField(StrFirstname.Get(wnd), firstname.Get()).InputValue(firstname).ErrorText(errFirstname.Get()).ID("onboarding-firstname").FullWidth(),
			ui.TextField(StrLastname.Get(wnd), lastname.Get()).InputValue(lastname).ErrorText(errLastname.Get()).ID("onboarding-lastname").FullWidth(),
			ui.PasswordField(StrPassword.Get(wnd), password.Get()).InputValue(password).ErrorText(errPassword.Get()).ID("onboarding-password").FullWidth(),
			ui.PasswordField(StrPasswordRepeated.Get(wnd), passwordRepeated.Get()).InputValue(passwordRepeated).ErrorText(errPassword.Get()).ID("onboarding-password-repeated").FullWidth(),
			uiuser.PasswordStrengthView(wnd, user.CalculatePasswordStrength(password.Get())),
		).Gap(ui.L8).FullWidth()

		actions = append(actions, ui.PrimaryButton(func() {
			errFirstname.Set(required(wnd, firstname.Get()))
			errLastname.Set(required(wnd, lastname.Get()))
			errPassword.Set("")
			if password.Get() != passwordRepeated.Get() {
				errPassword.Set(StrPasswordsDiffer.Get(wnd))
			}

			if errFirstname.Get() != "" || errLastname.Get() != "" || errPassword.Get() != "" {
				return
			}

			err := flow.Complete(onboarding.Profile{
				Firstname:        firstname.Get(),
				Lastname:         lastname.Get(),
				Password:         user.Password(password.Get()),
				PasswordRepeated: user.Password(passwordRepeated.Get()),
			})

			if errors.Is(err, onboarding.ErrNotVerified) {
				codeErr.Set(errorText(wnd, err))
				return
			}

			if err != nil {
				alert.ShowBannerError(wnd, err)
			}
		}).Title(StrCreateAccount.Get(wnd)).Key("onboarding", "create"))
	}

	return ui.VStack(
		ui.VStack(
			alert.BannerMessages(wnd),
			ui.WindowTitle(title),
			uisession.LoginRegisterCard(wnd, themeSettings.PageLogoLight, themeSettings.PageLogoDark, title, subtitle, body, actions...),
		).Gap(ui.L16).FullWidth(),
	).Frame(ui.Frame{}.MatchScreen())
}

func required(wnd core.Window, s string) string {
	if s == "" {
		return StrRequired.Get(wnd)
	}

	return ""
}

func errorText(wnd core.Window, err error) string {
	var cooldown onboarding.CooldownError
	switch {
	case errors.As(err, &cooldown):
		return StrErrCooldownX.Get(wnd, i18n.String("time", cooldown.RetryAt.In(time.Local).Format("15:04:05")))
	case errors.Is(err, onboarding.ErrInvalidCode):
		return StrErrInvalidCode.Get(wnd)
	case errors.Is(err, onboarding.ErrCodeExpired):
		return StrErrCodeExpired.Get(wnd)
	case errors.Is(err, onboarding.ErrTooManyAttempts):
		return StrErrTooManyAttempt.Get(wnd)
	case errors.Is(err, onboarding.ErrNoCode):
		return StrErrNoCode.Get(wnd)
	case errors.Is(err, onboarding.ErrNotVerified):
		return StrErrNotVerified.Get(wnd)
	default:
		return err.Error()
	}
}
