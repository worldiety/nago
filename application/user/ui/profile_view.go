// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/form"
	"strings"
)

func ViewProfile(wnd core.Window, roles []role.Role, groups []group.Group, licenses []license.UserLicense, email user.Email, contact user.Contact) core.View {
	var myRoleNames []string

	for _, myRole := range roles {
		myRoleNames = append(myRoleNames, myRole.Name)
	}

	if len(myRoleNames) == 0 {
		myRoleNames = append(myRoleNames, "Kein Rollenmitglied")
	}

	var myGroupNames []string

	for _, myGroup := range groups {
		myGroupNames = append(myGroupNames, myGroup.Name)
	}

	if len(myGroupNames) == 0 {
		myGroupNames = append(myGroupNames, "Kein Gruppenmitglied")
	}

	var myLicenseNames []string
	for _, l := range licenses {
		myLicenseNames = append(myLicenseNames, l.Name)
	}

	if len(myLicenseNames) == 0 {
		myLicenseNames = append(myLicenseNames, "Keine Lizenzen")
	}

	var avatarImg core.View
	if contact.Avatar == "" {
		avatarImg = avatar.Text(xstrings.Join2(" ", contact.Firstname, contact.Lastname)).Size(ui.L144)
	} else {
		avatarImg = avatar.URI(httpimage.URI(contact.Avatar, image.FitCover, 144, 144)).Size(ui.L144)
	}

	var tmpDetailsViews []core.View

	tmpDetailsViews = append(tmpDetailsViews, ui.Text(xstrings.Join2(" ", contact.Firstname, contact.Lastname)).Font(ui.SubTitle))
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
			ui.SecondaryButton(func() {
				core.HTTPOpen(wnd.Navigation(), core.URI("mailto:"+string(email)), "_self")
			}).AccessibilityLabel("eMail").PreIcon(heroSolid.Envelope),
		).Gap(ui.L8).Padding(ui.Padding{Top: ui.L4}))
	}

	contactDetails := ui.VStack(
		tmpDetailsViews...,
	).Alignment(ui.Leading)

	fakeState := core.StateOf[contactViewModel](wnd, string(email)+contact.Firstname+contact.Lastname).Init(func() contactViewModel {
		return newContactViewModel(string(email), contact)
	})

	return ui.VStack(
		ui.HStack(

			ui.VStack(
				ui.HStack(
					avatarImg,
					contactDetails,
				).Gap(ui.L20).Alignment(ui.Leading).FullWidth(),
				ui.HLine(),
				ui.If(contact.AboutMe != "", ui.Text(contact.AboutMe)),
				ui.If(contact.AboutMe != "", ui.HLine()),
				ui.Text(strings.Join(myRoleNames, ", ")).FullWidth().TextAlignment(ui.TextAlignEnd),
				ui.Text(strings.Join(myGroupNames, ", ")).FullWidth().TextAlignment(ui.TextAlignEnd),
				ui.Text(strings.Join(myLicenseNames, ", ")).FullWidth().TextAlignment(ui.TextAlignEnd),
				ui.Space(ui.L8),
				form.Auto(form.AutoOptions{
					SectionPadding: std.Some[ui.Padding](ui.Padding{}),
					ViewOnly:       true,
					IgnoreFields:   []string{"Avatar", "AboutMe"},
				}, fakeState),
			).Alignment(ui.Leading).
				FullWidth(),
		).Alignment(ui.Leading).
			FullWidth().
			Gap(ui.L20).
			BackgroundColor(ui.ColorCardBody).
			Border(ui.Border{}.Radius(ui.L16)),
	).FullWidth()

}
