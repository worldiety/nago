// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/admin"
	uiadmin "go.wdy.de/nago/application/admin/ui"
	uisession "go.wdy.de/nago/application/session/ui"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"log/slog"
)

type TAccountView struct {
	sections       []core.View
	uid            user.ID
	getDisplayName user.DisplayName
	wnd            core.Window
	logoutPage     core.NavigationPath
}

// AccountView requires a bunch of SystemService instances to work correctly:
//   - user.DisplayName
//   - uiuser.Pages
//   - uisession.Pages
func AccountView(wnd core.Window) TAccountView {
	getDisplayName, ok := core.SystemService[user.DisplayName](wnd.Application())
	if !ok {
		slog.Error("no system service user.DisplayName")
		getDisplayName = func(uid user.ID) user.Compact {
			return user.Compact{}
		}
	}

	var schemeModeIcon core.SVG
	var schemeModeText string
	if wnd.Info().ColorScheme == core.Light {
		schemeModeIcon = flowbiteOutline.Moon
		schemeModeText = "Dunkle Darstellung verwenden"
	} else {
		schemeModeIcon = flowbiteOutline.Sun
		schemeModeText = "Hello Darstellung verwenden"
	}

	userPages, _ := core.SystemService[Pages](wnd.Application())
	sessionPages, _ := core.SystemService[uisession.Pages](wnd.Application())

	c := TAccountView{
		wnd:            wnd,
		getDisplayName: getDisplayName,
		uid:            wnd.Subject().ID(),
		logoutPage:     sessionPages.Logout,
	}.Sections(
		AccountSection("Konto").
			Entries(
				AccountAction(flowbiteOutline.UserEdit, "Profil verwalten", func() {
					wnd.Navigation().ForwardTo(userPages.MyProfile, nil)
				}),
				AccountAction(schemeModeIcon, schemeModeText, func() {
					if wnd.Info().ColorScheme == core.Light {
						wnd.SetColorScheme(core.Dark)
					} else {
						wnd.SetColorScheme(core.Light)
					}

				}),
			),
	)

	if adminPages, ok := core.SystemService[uiadmin.Pages](wnd.Application()); ok {
		if queryGroups, ok := core.SystemService[admin.QueryGroups](wnd.Application()); ok {
			visibleEntries := queryGroups(wnd.Subject(), "")
			if len(visibleEntries) > 0 {
				c = c.WithSections(
					AccountSection("Einstellungen").
						Entries(
							AccountAction(flowbiteOutline.Cog, "Verwalten", func() {
								wnd.Navigation().ForwardTo(adminPages.AdminCenter, nil)
							}),
						),
				)
			}
		}

	}

	return c
}

func (c TAccountView) Sections(sections ...core.View) TAccountView {
	c.sections = sections
	return c
}

func (c TAccountView) WithSections(sections ...core.View) TAccountView {
	c.sections = append(c.sections, sections...)
	return c
}

func (c TAccountView) Render(ctx core.RenderContext) core.RenderNode {
	if !c.wnd.Subject().Valid() {
		return ui.Text("Kein Nutzer verfügbar").Render(ctx)
	}

	displayData := c.getDisplayName(c.uid)
	cfgTheme := core.GlobalSettings[theme.Settings](c.wnd)

	return ui.VStack(
		avatar.TextOrImage(displayData.Displayname, displayData.Avatar).Size(ui.L96),
		ui.VStack(
			ui.Text(displayData.Displayname).Font(ui.Large),
			ui.Text(string(displayData.Mail)),
		).Gap(ui.L4),
		ui.If(len(c.sections) > 0, ui.VStack(c.sections...).FullWidth().Gap(ui.L24)),
		ui.SecondaryButton(func() {
			if err := c.wnd.Logout(); err != nil {
				alert.ShowBannerError(c.wnd, err)
				return
			}

			c.wnd.Navigation().ForwardTo(c.logoutPage, nil)
		}).Title("Abmelden").Frame(ui.Frame{}.FullWidth()),
		ui.HLineWithColor(ui.ColorAccent),
		ui.HStack(
			ui.If(cfgTheme.Impress != "", ui.Link(c.wnd, "Impressum", cfgTheme.Impress, ui.LinkTargetNewWindowOrTab)),
			ui.If(cfgTheme.PrivacyPolicy != "", ui.Link(c.wnd, "Datenschutz", cfgTheme.PrivacyPolicy, ui.LinkTargetNewWindowOrTab)),
			ui.If(cfgTheme.GeneralTermsAndConditions != "", ui.Link(c.wnd, "AGB", cfgTheme.GeneralTermsAndConditions, ui.LinkTargetNewWindowOrTab)),
			ui.If(cfgTheme.TermsOfUse != "", ui.Link(c.wnd, "Nutzungsbedingungen", cfgTheme.TermsOfUse, ui.LinkTargetNewWindowOrTab)),
		).Gap(ui.L16),
	).Gap(ui.L16).
		FullWidth().
		Render(ctx)
}

type TAccountSection struct {
	name    string
	entries []core.View
}

func AccountSection(name string) TAccountSection {
	return TAccountSection{name: name}
}

func (c TAccountSection) Entries(entries ...core.View) TAccountSection {
	c.entries = entries
	return c
}

func (c TAccountSection) Render(ctx core.RenderContext) core.RenderNode {
	var tmp []core.View
	tmp = append(tmp, ui.Text(c.name).Font(ui.SubTitle).Padding(ui.Padding{Bottom: ui.L8}))
	tmp = append(tmp, c.entries...)

	return ui.VStack(
		tmp...,
	).Gap(ui.L4).
		FullWidth().
		Alignment(ui.Leading).
		Render(ctx)
}

func AccountAction(icon core.SVG, text string, action func()) core.View {
	return ui.HStack(
		ui.ImageIcon(icon),
		ui.Text(text).Font(ui.Font{Weight: ui.BoldFontWeight}),
		ui.Spacer(),
		ui.ImageIcon(flowbiteOutline.ChevronRight),
	).Gap(ui.L8).
		HoveredBackgroundColor(ui.I1).
		Action(action).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L4)).
		Border(ui.Border{}.Radius(ui.L8))
}
