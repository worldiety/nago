// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"errors"
	"go.wdy.de/nago/application/consent"
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
	consent user.Consent,
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
		actionCard(wnd, presentPasswordChange, findUserByID, consent),
	).Gap(ui.L20).
		Alignment(ui.Leading).
		Frame(ui.Frame{Width: ui.L560})
}

func passwordChangeDialog(wnd core.Window, changeMyPassword user.ChangeMyPassword, presentPasswordChange *core.State[bool]) core.View {
	if !presentPasswordChange.Get() {
		// security note: purge our states below, if dialog is not visible
		return nil
	}

	oldPassword := core.AutoState[string](wnd)
	password0 := core.AutoState[string](wnd)
	password1 := core.AutoState[string](wnd)
	errMsg := core.AutoState[error](wnd)
	oldPwdErrMsg := core.AutoState[string](wnd)
	newPwdErrMsg := core.AutoState[string](wnd)

	strength := user.CalculatePasswordStrength(password0.Get())
	body := ui.VStack(
		ui.If(errMsg.Get() != nil, ui.VStack(alert.BannerError(errMsg.Get())).Padding(ui.Padding{Bottom: ui.L20})),
		ui.PasswordField("Altes Passwort", oldPassword.Get()).
			ID("nago-password"). //this is the same ID as in package uilogin
			AutoComplete(false).
			InputValue(oldPassword).
			ErrorText(oldPwdErrMsg.Get()).
			Frame(ui.Frame{}.FullWidth()),

		ui.HLine(),
		ui.PasswordField("Neues Passwort", password0.Get()).
			AutoComplete(false).
			InputValue(password0).
			ErrorText(newPwdErrMsg.Get()).
			Frame(ui.Frame{}.FullWidth()),
		ui.Space(ui.L16),

		ui.PasswordField("Neues Passwort wiederholen", password1.Get()).
			AutoComplete(false).
			InputValue(password1).
			ErrorText(newPwdErrMsg.Get()).
			Frame(ui.Frame{}.FullWidth()),

		ui.Space(ui.L16),

		PasswordStrengthView(wnd, strength),
	).FullWidth()

	return alert.Dialog("Passwort ändern", body, presentPasswordChange, alert.Cancel(func() {
		errMsg.Set(nil)
		oldPassword.Set("")
		password0.Set("")
		password1.Set("")
	}),
		alert.MinWidth(ui.L560),
		alert.Custom(
			func(close func(closeDlg bool)) core.View {
				return ui.PrimaryButton(func() {
					errMsg.Set(nil)
					oldPwdErrMsg.Set("")
					newPwdErrMsg.Set("")

					if err := changeMyPassword(wnd.Subject(), user.Password(oldPassword.Get()), user.Password(password0.Get()), user.Password(password1.Get())); err != nil {

						switch {
						case errors.Is(err, user.NewPasswordMustBeDifferentFromOldPasswordErr):
							newPwdErrMsg.Set("Das alte und das neue Kennwort müssen sich unterscheiden.")
						case errors.Is(err, user.PasswordsDontMatchErr):
							newPwdErrMsg.Set("Die Passwörter stimmen nicht überein")
						case errors.Is(err, user.InvalidOldPasswordErr):
							oldPwdErrMsg.Set("Das alte Kennwort ist falsch.")
						default:
							errMsg.Set(err)
						}

						return
					}

					// security note: purge passwords from memory
					oldPassword.Set("")
					password0.Set("")
					password1.Set("")

					close(true)
				}).Enabled(strength.Acceptable).Title("Passwort ändern")

			},
		))
}

func actionCard(wnd core.Window, presentPasswordChange *core.State[bool], findUserByID user.FindByID, consentFn user.Consent) core.View {
	cfgUsers := core.GlobalSettings[user.Settings](wnd)

	optUsr, err := findUserByID(wnd.Subject(), wnd.Subject().ID())
	if err != nil {
		alert.ShowBannerError(wnd, err)
	}

	if optUsr.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	usr := optUsr.Unwrap()
	consents := usr.CompatConsents()

	var actionItems []core.View
	for _, consentOption := range cfgUsers.Consents {
		if consentOption.Required {
			// not clear, how to handle this. Some information must never be changed or confirmed again, like
			// min age. But others, like changed terms and conditions, must be approved again.
			continue
		}

		acceptedState := core.StateOf[bool](wnd, string(consentOption.ID)).Init(func() bool {
			return consent.HasApproved(consents, consentOption.ID)
		}).Observe(func(newValue bool) {
			var status consent.Status
			if newValue {
				status = consent.Approved
			}

			action := consent.Action{
				Location: string(wnd.Path()),
				Status:   status,
			}

			if err := consentFn(wnd.Subject(), wnd.Subject().ID(), consentOption.ID, action); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
		})

		actionItems = append(actionItems, list.Entry().
			Headline(consentOption.Profile.Label).
			SupportingText(consentOption.Profile.SupportingText).
			Trailing(ui.Toggle(acceptedState.Get()).InputChecked(acceptedState)))
	}

	actionItems = append(actionItems, list.Entry().
		Headline("Passwort ändern").
		Action(func() {
			presentPasswordChange.Set(true)
		}).
		Frame(ui.Frame{Height: ui.L48}.FullWidth()).
		Trailing(ui.ImageIcon(heroSolid.ChevronRight)))

	return list.List(actionItems...).Frame(ui.Frame{}.FullWidth())
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
