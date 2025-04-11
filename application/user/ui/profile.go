// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/image"
	httpimage "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
	"os"
	"strings"
)

func ProfilePage(
	wnd core.Window,
	pages Pages,
	changeMyPassword user.ChangeMyPassword,
	readMyContact user.ReadMyContact,
	findMyRoles role.FindMyRoles,
	findUserByID user.FindByID,
	adoptNewsletter user.AdoptNewsletter,
	adoptSMS user.AdoptSMS,
) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(user.PermissionDeniedErr)
	}

	contact, err := readMyContact(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}

	presentPasswordChange := core.AutoState[bool](wnd)
	return ui.VStack(
		passwordChangeDialog(wnd, changeMyPassword, presentPasswordChange),
		ui.H1("Mein Profil"),
		profileCard(wnd, pages, contact, findMyRoles),
		actionCard(wnd, presentPasswordChange, findUserByID, adoptNewsletter, adoptSMS),
	).Gap(ui.L20).
		Alignment(ui.Leading).
		Frame(ui.Frame{Width: ui.L560})
}

func passwordChangeDialog(wnd core.Window, changeMyPassword user.ChangeMyPassword, presentPasswordChange *core.State[bool]) core.View {
	oldPassword := core.AutoState[string](wnd)
	password0 := core.AutoState[string](wnd)
	password1 := core.AutoState[string](wnd)
	errMsg := core.AutoState[error](wnd)
	body := ui.VStack(
		ui.If(errMsg.Get() != nil, ui.VStack(alert.BannerError(errMsg.Get())).Padding(ui.Padding{Bottom: ui.L20})),
		ui.PasswordField("Altes Passwort", oldPassword.Get()).
			AutoComplete(false).
			InputValue(oldPassword).
			Frame(ui.Frame{}.FullWidth()),
		ui.HLine(),
		ui.PasswordField("Neues Passwort", password0.Get()).
			AutoComplete(false).
			InputValue(password0).
			Frame(ui.Frame{}.FullWidth()),
		ui.PasswordField("Neues Passwort wiederholen", password1.Get()).
			AutoComplete(false).
			InputValue(password1).
			Frame(ui.Frame{}.FullWidth()),
	).FullWidth()

	return alert.Dialog("Passwort ändern", body, presentPasswordChange, alert.Cancel(func() {
		errMsg.Set(nil)
		oldPassword.Set("")
		password0.Set("")
		password1.Set("")
	}), alert.Save(func() (close bool) {
		if err := changeMyPassword(wnd.Subject(), user.Password(oldPassword.Get()), user.Password(password0.Get()), user.Password(password1.Get())); err != nil {
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

func actionCard(wnd core.Window, presentPasswordChange *core.State[bool], findUserByID user.FindByID, adoptNewsletter user.AdoptNewsletter, adoptSMS user.AdoptSMS) core.View {
	cfgUsers := core.GlobalSettings[user.Settings](wnd)

	var usr user.User
	if cfgUsers.CanReceiveSMS || cfgUsers.CanAcceptNewsletter {
		optUsr, err := findUserByID(wnd.Subject(), wnd.Subject().ID())
		if err != nil {
			alert.ShowBannerError(wnd, err)
		}

		if optUsr.IsNone() {
			alert.ShowBannerError(wnd, os.ErrNotExist)
		}

		if optUsr.IsSome() {
			usr = optUsr.Unwrap()
		}
	}

	return list.List(
		ui.IfFunc(cfgUsers.CanAcceptNewsletter, func() core.View {
			acceptedState := core.AutoState[bool](wnd).Init(func() bool {
				return !usr.Newsletter.ApprovedAt.IsZero()
			}).Observe(func(newValue bool) {
				if err := adoptNewsletter(wnd.Subject(), wnd.Subject().ID(), newValue); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
			})

			return list.Entry().
				Headline("Newsletter erhalten").
				SupportingText("Neuigkeiten erhalten und über Änderungen informiert werden.").
				Trailing(ui.Toggle(acceptedState.Get()).InputChecked(acceptedState))
		}),

		ui.IfFunc(cfgUsers.CanReceiveSMS, func() core.View {
			acceptedState := core.AutoState[bool](wnd).Init(func() bool {
				return !usr.SMS.ApprovedAt.IsZero()
			}).Observe(func(newValue bool) {
				if err := adoptSMS(wnd.Subject(), wnd.Subject().ID(), newValue); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
			})

			return list.Entry().
				Headline("SMS erhalten").
				SupportingText("Kurzfristige Aktualisierungen und Benachrichtigungen erhalten.").
				Trailing(ui.Toggle(acceptedState.Get()).InputChecked(acceptedState))
		}),

		list.Entry().
			Headline("Passwort ändern").
			Action(func() {
				presentPasswordChange.Set(true)
			}).
			Frame(ui.Frame{Height: ui.L48}.FullWidth()).
			Trailing(ui.ImageIcon(heroSolid.ChevronRight)),
	).Frame(ui.Frame{}.FullWidth())
}

func profileCard(wnd core.Window, pages Pages, contact user.Contact, findMyRoles role.FindMyRoles) core.View {
	var myRoleNames []string
	for myRole, err := range findMyRoles(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		myRoleNames = append(myRoleNames, myRole.Name)
	}

	if len(myRoleNames) == 0 {
		myRoleNames = append(myRoleNames, "Kein Rollenmitglied")
	}

	var avatarImg core.View
	if contact.Avatar == "" {
		avatarImg = avatar.Text(wnd.Subject().Name()).Size(ui.L144)
	} else {
		avatarImg = avatar.URI(httpimage.URI(contact.Avatar, image.FitCover, 144, 144)).Size(ui.L144)
	}

	var tmpDetailsViews []core.View

	tmpDetailsViews = append(tmpDetailsViews, ui.Text(wnd.Subject().Name()).Font(ui.SubTitle))
	if contact.Position != "" {
		tmpDetailsViews = append(tmpDetailsViews, ui.Text(contact.Position))
	}

	if contact.CompanyName != "" {
		tmpDetailsViews = append(tmpDetailsViews, ui.Text(contact.CompanyName))
	}

	if adr := xstrings.Join2(" ", contact.PostalCode, contact.City); adr != "" {
		tmpDetailsViews = append(tmpDetailsViews, ui.Text(adr))
	}

	if contact.LinkedIn != "" || contact.Website != "" {
		tmpDetailsViews = append(tmpDetailsViews, ui.HStack(
			ui.If(contact.LinkedIn != "", ui.SecondaryButton(func() {
				core.HTTPOpen(wnd.Navigation(), core.HTTPify(contact.LinkedIn), "_blank")
			}).AccessibilityLabel("LinkedIn").
				PreIcon(flowbiteSolid.Linkedin)),
			ui.If(contact.Website != "", ui.SecondaryButton(func() {
				core.HTTPOpen(wnd.Navigation(), core.HTTPify(contact.Website), "_blank")
			}).AccessibilityLabel("Webseite").
				PreIcon(heroSolid.GlobeEuropeAfrica)),
		).Gap(ui.L8).Padding(ui.Padding{Top: ui.L4}))
	}

	contactDetails := ui.VStack(
		tmpDetailsViews...,
	).Alignment(ui.Leading)

	return ui.VStack(
		ui.HStack(
			ui.Text(strings.Join(myRoleNames, ", "))).
			FullWidth().
			Alignment(ui.Leading).
			BackgroundColor(ui.ColorCardTop).
			Padding(ui.Padding{}.Horizontal(ui.L20).Vertical(ui.L12)),
		ui.VStack(
			ui.HStack(
				avatarImg,
				contactDetails,
			).Gap(ui.L20),
			ui.HLineWithColor(ui.ColorAccent),
			ui.HStack(
				ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo(pages.MyContact, nil)
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
