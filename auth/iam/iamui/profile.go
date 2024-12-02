package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
)

func ProfilePage(wnd core.Window, changeMyPassword iam.ChangeMyPassword) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(auth.NotLoggedIn(""))
	}

	presentPasswordChange := core.AutoState[bool](wnd)
	return ui.VStack(
		passwordChangeDialog(wnd, changeMyPassword, presentPasswordChange),
		ui.H1("Mein Profil"),
		profileCard(wnd),
		actionCard(wnd, presentPasswordChange),
	).Gap(ui.L20).
		Alignment(ui.Leading).
		Frame(ui.Frame{Width: ui.L560})
}

func passwordChangeDialog(wnd core.Window, changeMyPassword iam.ChangeMyPassword, presentPasswordChange *core.State[bool]) core.View {
	oldPassword := core.AutoState[string](wnd)
	password0 := core.AutoState[string](wnd)
	password1 := core.AutoState[string](wnd)
	errMsg := core.AutoState[error](wnd)
	body := ui.VStack(
		ui.If(errMsg.Get() != nil, ui.VStack(alert.BannerError(errMsg.Get())).Padding(ui.Padding{Bottom: ui.L20})),
		ui.PasswordField("Altes Passwort", oldPassword.Get()).InputValue(oldPassword).Frame(ui.Frame{}.FullWidth()),
		ui.HLine(),
		ui.PasswordField("Neues Passwort", password0.Get()).InputValue(password0).Frame(ui.Frame{}.FullWidth()),
		ui.PasswordField("Neues Passwort wiederholen", password1.Get()).InputValue(password1).Frame(ui.Frame{}.FullWidth()),
	).FullWidth()
	return alert.Dialog("Passwort ändern", body, presentPasswordChange, alert.Cancel(func() {
		errMsg.Set(nil)
		oldPassword.Set("")
		password0.Set("")
		password1.Set("")
	}), alert.Save(func() (close bool) {
		if err := changeMyPassword(wnd.Subject(), iam.Password(oldPassword.Get()), iam.Password(password0.Get()), iam.Password(password1.Get())); err != nil {
			errMsg.Set(err)
			return false
		}

		errMsg.Set(nil)
		oldPassword.Set("")
		password0.Set("")
		password1.Set("")

		return true
	}))
}

func actionCard(wnd core.Window, presentPasswordChange *core.State[bool]) core.View {
	return list.List(
		list.Entry().
			Headline("Passwort ändern").
			Action(func() {
				presentPasswordChange.Set(true)
			}).
			Frame(ui.Frame{Height: ui.L48}.FullWidth()).
			Trailing(ui.ImageIcon(heroSolid.ChevronRight)),
	).Frame(ui.Frame{}.FullWidth())
}

func profileCard(wnd core.Window) core.View {
	var firstRole auth.RID
	for rid := range wnd.Subject().Roles() {
		firstRole = rid
		break
	}

	return ui.VStack(
		ui.HStack(
			ui.Text(string(firstRole))).
			FullWidth().
			Alignment(ui.Leading).
			BackgroundColor(ui.ColorCardTop).
			Padding(ui.Padding{}.Horizontal(ui.L20).Vertical(ui.L12)),
		ui.VStack(
			ui.HStack(
				avatar.Text(wnd.Subject().Name()).Size(ui.L144),
				ui.Text(wnd.Subject().Name()).Font(ui.SubTitle),
			).Gap(ui.L20),
			ui.HLineWithColor(ui.ColorAccent),
			ui.HStack(
				ui.SecondaryButton(func() {

				}).Title("Bearbeiten"),
			).Alignment(ui.Trailing).
				FullWidth(),
		).Alignment(ui.Leading).
			FullWidth().
			Padding(ui.Padding{Bottom: ui.L20}.Horizontal(ui.L20)),
	).Alignment(ui.Leading).
		FullWidth().
		Gap(ui.L20).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L16))
}
