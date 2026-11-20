// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"log/slog"

	"go.wdy.de/nago/application/admin"
	uiadmin "go.wdy.de/nago/application/admin/ui"
	uisession "go.wdy.de/nago/application/session/ui"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	icons "go.wdy.de/nago/presentation/icons/material/outlined"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
)

var (
	StrSwitchTheme = core.DefaultStr("account-view.switch-theme", "Switch theme", "Farbschema wechseln")
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
//   - admin.HasEntries and uiadmin.Pages, to link the admin center
//
// It shows nago's sections (see AccountSectionModel) as changed by the AccountSections of the context, which
// application.Configurator.AddAccountSections registers.
func AccountView(wnd core.Window) TAccountView {
	getDisplayName, ok := core.FromContext[user.DisplayName](wnd.Context(), "")
	if !ok {
		slog.Error("no system service user.DisplayName")
		getDisplayName = func(uid user.ID) user.Compact {
			return user.Compact{ID: uid}
		}
	}

	sessionPages, _ := core.FromContext[uisession.Pages](wnd.Context(), "")

	sections := accountSections(wnd)
	if fn, ok := core.FromContext[AccountSections](wnd.Context(), ""); ok && fn != nil {
		sections = fn(wnd, sections)
	}

	var views []core.View
	for _, section := range sections {
		if len(section.Entries) == 0 {
			continue
		}

		entries := make([]core.View, 0, len(section.Entries))
		for _, entry := range section.Entries {
			entries = append(entries, entry.View)
		}

		views = append(views, AccountSection(section.Title).ID(string(section.ID)).Entries(entries...))
	}

	return TAccountView{
		wnd:            wnd,
		getDisplayName: getDisplayName,
		uid:            wnd.Subject().ID(),
		logoutPage:     sessionPages.Logout,
	}.Sections(views...)
}

// accountSections are nago's sections of the account view of wnd: SectionAccount with EntryProfile and EntryTheme,
// then SectionSettings with EntryAdminCenter, if admin.HasEntries holds for the subject.
func accountSections(wnd core.Window) []AccountSectionModel {
	userPages, _ := core.FromContext[Pages](wnd.Context(), "")

	sections := []AccountSectionModel{
		{
			ID:    SectionAccount,
			Title: "Konto",
			Entries: []AccountEntryModel{
				{
					ID: EntryProfile,
					View: AccountAction(flowbiteOutline.UserEdit, "Profil verwalten", func() {
						wnd.Navigation().ForwardTo(userPages.MyProfile, nil)
					}),
				},
				{ID: EntryTheme, View: switchThemeButton(wnd)},
			},
		},
	}

	if adminPages, ok := core.FromContext[uiadmin.Pages](wnd.Context(), ""); ok {
		if hasEntries, ok := core.FromContext[admin.HasEntries](wnd.Context(), ""); ok && hasEntries(wnd.Subject()) {
			sections = append(sections, AccountSectionModel{
				ID:    SectionSettings,
				Title: "Einstellungen",
				Entries: []AccountEntryModel{
					{
						ID: EntryAdminCenter,
						View: AccountAction(flowbiteOutline.Cog, "Verwalten", func() {
							wnd.Navigation().ForwardTo(adminPages.AdminCenter, nil)
						}),
					},
				},
			})
		}
	}

	return sections
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
	hasLegalTexts := cfgTheme.Impress != "" || cfgTheme.PrivacyPolicy != "" || cfgTheme.GeneralTermsAndConditions != "" || cfgTheme.TermsOfUse != ""

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
		// the line and the legal texts only with at least one legal text, otherwise an empty line would remain
		ui.If(hasLegalTexts, ui.HLineWithColor(ui.ColorAccent)),
		ui.If(hasLegalTexts, ui.HStack(
			ui.If(cfgTheme.Impress != "", ui.Link(c.wnd, "Impressum", cfgTheme.Impress, ui.LinkTargetNewWindowOrTab)),
			ui.If(cfgTheme.PrivacyPolicy != "", ui.Link(c.wnd, "Datenschutz", cfgTheme.PrivacyPolicy, ui.LinkTargetNewWindowOrTab)),
			ui.If(cfgTheme.GeneralTermsAndConditions != "", ui.Link(c.wnd, "AGB", cfgTheme.GeneralTermsAndConditions, ui.LinkTargetNewWindowOrTab)),
			ui.If(cfgTheme.TermsOfUse != "", ui.Link(c.wnd, "Nutzungsbedingungen", cfgTheme.TermsOfUse, ui.LinkTargetNewWindowOrTab)),
		).Gap(ui.L16)),
	).Gap(ui.L16).
		FullWidth().
		Render(ctx)
}

type TAccountSection struct {
	id      string
	name    string
	entries []core.View
}

func AccountSection(name string) TAccountSection {
	return TAccountSection{name: name}
}

// ID sets the id of the rendered section.
func (c TAccountSection) ID(id string) TAccountSection {
	c.id = id
	return c
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
		ID(c.id).
		FullWidth().
		Alignment(ui.Leading).
		Render(ctx)
}

func AccountAction(icon core.SVG, text string, action func()) core.View {
	return ui.HStack(
		ui.ImageIcon(icon),
		ui.Text(text).Font(ui.Font{Weight: ui.HeadlineAndTitleFontWeight}),
		ui.Spacer(),
		ui.ImageIcon(flowbiteOutline.ChevronRight),
	).Gap(ui.L8).
		HoveredBackgroundColor(ui.I1).
		Action(action).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L4)).
		Border(ui.Border{}.Radius(ui.L8))
}

func switchThemeButton(wnd core.Window) core.View {
	return ui.ThemeSwitcher(
		ui.HStack(
			ui.ImageIcon(icons.Palette).FillColor(ui.M8),
			ui.Text(StrSwitchTheme.Get(wnd)).Font(ui.Font{Weight: ui.HeadlineAndTitleFontWeight}),
			ui.Spacer(),
			ui.ImageIcon(flowbiteOutline.ChevronRight),
		).Gap(ui.L8).
			HoveredBackgroundColor(ui.I1).
			FullWidth().
			Padding(ui.Padding{}.All(ui.L4)).
			Border(ui.Border{}.Radius(ui.L8)),
	).Frame(ui.Frame{}.FullWidth())
}
