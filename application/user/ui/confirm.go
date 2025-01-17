package uiuser

import (
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

type SendVerificationMail func(uid user.ID) error

func ConfirmPage(
	wnd core.Window,
	loginPage core.NavigationPath,
	confirmMail user.ConfirmMail,
	sendMail SendVerificationMail,
	requiresPwdChange user.RequiresPasswordChange,
	sysUser user.SysUser,
	setPassword user.ChangeOtherPassword,
	logoutFn session.Logout,
) core.View {
	uid := user.ID(wnd.Values()["id"])
	code := wnd.Values()["code"]

	presented := core.AutoState[bool](wnd)
	confirmed := core.AutoState[bool](wnd)

	if wnd.Subject().Valid() && uid == wnd.Subject().ID() {
		return ui.VStack(
			ui.VStack(
				ui.H1("Konto verifizieren"),
				ui.Text("Das Konto ist verifiziert und aktuell. Es gibt nichts zu tun."),
			).Alignment(ui.Leading).
				Frame(ui.Frame{MaxWidth: ui.L560}),
		).Gap(ui.L16).FullWidth()
	}

	if !confirmed.Get() {
		errState := core.AutoState[error](wnd)
		// security note: hide the process behind a button, because otherwise mail client
		// which apply a preview, will verify just by invoking the endpoint
		return ui.VStack(
			ui.H1("Konto verifizieren"),
			alert.Dialog("E-Mail gesendet", ui.Text("Eine neue E-Mail mit einem aktualisierten Link wurde an Ihr Postfach gesendet. Schließen Sie diese Seite und klicken Sie auf den neuen Link aus der E-Mail."), presented),
			alert.BannerError(errState.Get()),
			ui.IfFunc(errState.Get() != nil, func() core.View {
				return ui.PrimaryButton(func() {
					if err := sendMail(uid); err != nil {
						alert.ShowBannerError(wnd, err)
					}

					presented.Set(true)
				}).Title("Neuen Verifikationslink anfragen")
			}),
			ui.IfFunc(errState.Get() == nil, func() core.View {
				return ui.PrimaryButton(func() {
					errState.Set(confirmMail(uid, code))
					if errState.Get() == nil {
						// security note: as long as we ever reached this, we can be confident, that the first time we had success
						// is tracked through each render cycle
						confirmed.Set(true)
					}
				}).Title("Konto jetzt bestätigen")
			}),
		).Gap(ui.L16)

	}

	requiresChange, err := requiresPwdChange(uid)
	if err != nil {
		return ui.VStack(alert.BannerError(err))
	}

	pwd0 := core.AutoState[string](wnd)
	pwd1 := core.AutoState[string](wnd)
	pwdErr := core.AutoState[error](wnd)
	invalidate := core.AutoState[int](wnd)

	var body core.View
	if requiresChange {
		if wnd.Subject().Valid() && wnd.Subject().ID() != uid {
			// security note: even though this is not necessary, it looks wrong
			// from the user perspective, thus be clear and logout if we are
			// talking about different users

			if _, err := logoutFn(wnd.Session().ID()); err != nil {
				return alert.BannerError(err)
			}

			wnd.UpdateSubject(nil)
		}

		body = ui.VStack(
			ui.Text("Das Konto wurde bestätigt aber das Kennwort muss noch geändert werden."),
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
				if err := setPassword(sysUser(), uid, user.Password(pwd0.Get()), user.Password(pwd1.Get())); err != nil {
					pwdErr.Set(err)
					return
				}

				pwdErr.Set(nil)

				// we changed the backend, thus trigger a new render cycle, the pwdErr may not trigger if still nil
				invalidate.Set(invalidate.Get() + 1)
			}).Title("Kennwort aktualisieren"),
		).Gap(ui.L16)
	} else {
		body = ui.VStack(
			ui.Text("Das Konto wurde bestätigt und ist aktuell."),
			ui.PrimaryButton(func() {
				wnd.Navigation().ResetTo(loginPage, nil)
			}).Title("Jetzt anmelden"),
		).Gap(ui.L16)
	}

	return ui.VStack(
		ui.VStack(
			ui.H1("Konto verifizieren"),
			body,
		).Alignment(ui.Center).
			Frame(ui.Frame{MaxWidth: ui.L560}),
	).Gap(ui.L16).FullWidth()

}
