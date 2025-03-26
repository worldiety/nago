// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func ResetPasswordPage(
	wnd core.Window,
	loginPage core.NavigationPath,
	changePassword user.ChangePasswordWithCode,
	logoutFn session.Logout,
) core.View {
	uid := user.ID(wnd.Values()["id"])
	code := wnd.Values()["code"]

	requiresChange := core.AutoState[bool](wnd).Init(func() bool {
		return true
	})

	pwd0 := core.AutoState[string](wnd)
	pwd1 := core.AutoState[string](wnd)
	pwdErr := core.AutoState[error](wnd)

	var body core.View
	if requiresChange.Get() {

		body = ui.VStack(
			ui.IfFunc(pwdErr.Get() != nil, func() core.View {
				return alert.BannerError(pwdErr.Get())
			}),
			ui.PasswordField("Neues Kennwort", pwd0.Get()).
				AutoComplete(false).
				InputValue(pwd0),
			ui.PasswordField("Kennwort wiederholen", pwd1.Get()).
				AutoComplete(false).
				InputValue(pwd1),
			ui.PrimaryButton(func() {
				if err := changePassword(uid, code, user.Password(pwd0.Get()), user.Password(pwd1.Get())); err != nil {
					pwdErr.Set(err)
					return
				}

				if wnd.Subject().Valid() {
					// security note: even though this is not necessary, it looks wrong
					// from the user perspective, thus be clear and logout to be clear

					if _, err := logoutFn(wnd.Session().ID()); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					wnd.UpdateSubject(nil)
				}

				pwdErr.Set(nil)

				requiresChange.Set(false)
			}).Title("Kennwort aktualisieren"),
		).Gap(ui.L16)
	} else {
		body = ui.VStack(
			ui.Text("Das Kennwort wurde erfolgreich geändert."),
			ui.PrimaryButton(func() {
				wnd.Navigation().ResetTo(loginPage, nil)
			}).Title("Jetzt anmelden"),
		).Gap(ui.L16)
	}

	return ui.VStack(
		ui.VStack(
			ui.H1("Kennwort zurücksetzen"),
			body,
		).Alignment(ui.Center).
			Frame(ui.Frame{MaxWidth: ui.L560}),
	).Gap(ui.L16).FullWidth()

}
