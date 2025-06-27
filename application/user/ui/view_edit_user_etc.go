// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"errors"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"time"
)

func viewEtc(wnd core.Window, ucUsers user.UseCases, usr *core.State[user.User]) core.View {
	presentedPwdChange := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.Text("Die Änderungen und Aktionen werden sofort angewendet und können nicht durch 'Abbrechen' rückgängig gemacht werden."),
		etcAction(
			wnd,
			"Nutzer über Konto benachrichtigen",
			"Den Nutzer per E-Mail darüber benachrichtigen, dass dieses Konto angelegt wurde und ihn auffordern sich anzumelden. Dazu wird das gleiche interne Domänen-Ereignis erzeugt, als ob dieser Nutzer neu angelegt wurde. Prozesse oder Abläufe die davon ausgehen, dass dieses Ereignis einmalig ist, können sich womöglich fehlerhaft verhalten.",
			"",
			"Nutzer benachrichtigen",
			func() {
				bus := wnd.Application().EventBus()
				user.PublishUserCreated(bus, usr.Get(), true)
				alert.ShowBannerMessage(wnd, alert.Message{
					Title:    "Nutzer erstellt",
					Message:  "Ereignis für " + usr.String() + " erstellt.",
					Intent:   alert.IntentOk,
					Duration: time.Second * 2,
				})
			},
		),

		etcAction(
			wnd,
			"Nutzer löschen",
			"Den Nutzer aus dem System unwiderruflich entfernen. Einige mit dem Nutzer verbundene Ressourcen oder getrennt gespeicherte personenbezogenen Daten bleiben möglicherweise weiterhin im System erhalten.",
			"Den Nutzer wirklich löschen? Dieser Vorgang kann nicht rückgängig gemacht werden.",
			"Nutzer löschen",
			func() {
				if err := ucUsers.Delete(wnd.Subject(), usr.Get().ID); err != nil {
					alert.ShowBannerError(wnd, err)
				}
			},
		),

		ui.If(usr.Get().Enabled(),
			etcAction(
				wnd,
				"Nutzer deaktivieren",
				"Das Konto des Nutzer ist derzeit aktiv. Damit kann sich der Nutzer mit Kenntnis der E-Mail und des Kennwortes am System anmelden. Wenn das Konto deaktiviert wird, wird der Nutzer innerhalb von spätestens 5 Minuten automatisch vom System abgemeldet. Weitere Anmeldungen sind dann nicht mehr möglich.",
				"",
				"Nutzer deaktivieren",
				func() {
					if err := ucUsers.UpdateAccountStatus(wnd.Subject(), usr.Get().ID, user.Disabled{}); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					alert.ShowBannerMessage(wnd, alert.Message{
						Title:   "Nutzer deaktiviert",
						Message: "Der Nutzer " + usr.String() + " wurde deaktiviert.",
						Intent:  alert.IntentOk,
					})
				},
			),
		),

		ui.If(!usr.Get().Enabled(),
			etcAction(
				wnd,
				"Nutzer aktivieren",
				"Das Konto des Nutzer ist derzeit inaktiv. Daher kann sich der Nutzer derzeit nicht anmelden. Wenn das Konto wieder aktiviert wird, kann der Nutzer sich wieder anmelden bzw. ggf. sein Kennwort setzen oder das Konto per Double-Opt-In freischalten, falls erforderlich.",
				"",
				"Nutzer aktivieren",
				func() {
					if err := ucUsers.UpdateAccountStatus(wnd.Subject(), usr.Get().ID, user.Enabled{}); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					alert.ShowBannerMessage(wnd, alert.Message{
						Title:   "Nutzer aktiviert",
						Message: "Der Nutzer " + usr.String() + " wurde aktiviert.",
						Intent:  alert.IntentOk,
					})
				},
			),
		),

		ui.If(!usr.Get().VerificationCode.IsZero(),
			etcAction(
				wnd,
				"E-Mail bestätigen",
				"Das Konto des Nutzer ist derzeit noch nicht bestätigt. Daher kann sich der Nutzer derzeit nicht anmelden. Eine manuelle Bestätigung der E-Mail sollte nur erfolgen, wenn über einen sicheren Kanal bestätigt wurde, dass die Person, welche die Anmeldedaten besitzt auch wirklich der Inhaber der E-Mail Adresse ist. Achtung: Telefonnummern, Stimmen und Videoanrufe können gefälscht sein.",
				"",
				"E-Mail bestätigen",
				func() {
					if err := ucUsers.UpdateVerification(wnd.Subject(), usr.Get().ID, true); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					alert.ShowBannerMessage(wnd, alert.Message{
						Title:   "Nutzer bestätigt",
						Message: "Der Nutzer " + usr.String() + " wurde bestätigt.",
						Intent:  alert.IntentOk,
					})
				},
			),
		),

		passwordChangeOtherDialog(wnd, usr.Get().ID, ucUsers.ChangeOtherPassword, presentedPwdChange),
		etcAction(
			wnd,
			"Kennwort setzen",
			"Das Kennwort sollte aus Sicherheitsgründen immer vom Konto-Inhaber selbst gesetzt werden. Fall dies nicht möglich sein sollte, kann das Kennwort hier manuell vergeben werden. Es muss über einen sicheren Kanal bestätigt werden, dass die Person, welche die Anmeldedaten besitzt auch wirklich der Kontoinhaber ist. Achtung: Telefonnummern, Stimmen und Videoanrufe können gefälscht sein.",
			"",
			"Kennwort setzen",
			func() {
				presentedPwdChange.Set(true)
			},
		),
	).FullWidth().Gap(ui.L32)
}

func etcAction(wnd core.Window, title, text, confirmText, actionText string, action func()) core.View {
	presented := core.StateOf[bool](wnd, title)
	return ui.VStack(
		func() core.View {
			if !presented.Get() {
				return nil
			}

			return alert.Dialog("Aktion bestätigen", ui.Text(confirmText), presented, alert.Cancel(nil), alert.Confirm(func() (close bool) {
				action()
				return true
			}))

		}(),
		ui.VStack(
			ui.H2(title),
			ui.Text(text),
			ui.HStack(
				ui.SecondaryButton(func() {
					if confirmText == "" {
						action()
					} else {
						presented.Set(true)
					}
				}).Title(actionText),
			).FullWidth().Alignment(ui.Trailing),
		).FullWidth().Alignment(ui.Leading).Gap(ui.L8).Border(ui.Border{}.Radius(ui.L16).Width(ui.L1).Color(ui.ColorInputBorder)).Padding(ui.Padding{}.All(ui.L16)),
	).FullWidth().Alignment(ui.Leading).Gap(ui.L32)
}

func passwordChangeOtherDialog(wnd core.Window, uid user.ID, changeOtherPassword user.ChangeOtherPassword, presentPasswordChange *core.State[bool]) core.View {
	if !presentPasswordChange.Get() {
		// security note: purge our states below, if dialog is not visible
		return nil
	}

	password0 := core.AutoState[string](wnd)
	password1 := core.AutoState[string](wnd)
	errMsg := core.AutoState[error](wnd)
	newPwdErrMsg := core.AutoState[string](wnd)

	strength := user.CalculatePasswordStrength(password0.Get())
	body := ui.VStack(
		ui.If(errMsg.Get() != nil, ui.VStack(alert.BannerError(errMsg.Get())).Padding(ui.Padding{Bottom: ui.L20})),

		ui.HLine(),
		ui.PasswordField("Neues Passwort", password0.Get()).
			ID("other-new-pwd").
			AutoComplete(false).
			InputValue(password0).
			ErrorText(newPwdErrMsg.Get()).
			Frame(ui.Frame{}.FullWidth()),
		ui.Space(ui.L16),

		ui.PasswordField("Neues Passwort wiederholen", password1.Get()).
			AutoComplete(false).
			ID("other-new-pwd-repeat").
			InputValue(password1).
			ErrorText(newPwdErrMsg.Get()).
			Frame(ui.Frame{}.FullWidth()),

		ui.Space(ui.L16),

		PasswordStrengthView(wnd, strength),
	).FullWidth()

	return alert.Dialog("Passwort ändern", body, presentPasswordChange, alert.Cancel(func() {
		errMsg.Set(nil)
		password0.Set("")
		password1.Set("")
	}),
		alert.Width(ui.L560),
		alert.Custom(
			func(close func(closeDlg bool)) core.View {
				return ui.PrimaryButton(func() {
					errMsg.Set(nil)
					newPwdErrMsg.Set("")

					if err := changeOtherPassword(wnd.Subject(), uid, user.Password(password0.Get()), user.Password(password1.Get())); err != nil {

						switch {
						case errors.Is(err, user.NewPasswordMustBeDifferentFromOldPasswordErr):
							newPwdErrMsg.Set("Das alte und das neue Kennwort müssen sich unterscheiden.")
						case errors.Is(err, user.PasswordsDontMatchErr):
							newPwdErrMsg.Set("Die Passwörter stimmen nicht überein")
						default:
							errMsg.Set(err)
						}

						return
					}

					// security note: purge passwords from memory
					password0.Set("")
					password1.Set("")

					close(true)
				}).Enabled(strength.Acceptable).Title("Passwort ändern")

			},
		))
}
