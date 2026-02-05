// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	_ "embed"
	"slices"
	"strings"

	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/theme"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/auth"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui/footer"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/tracking"
)

type MenuEntryBuilder struct {
	parent               *ScaffoldBuilder
	icon                 core.SVG
	customView           core.View
	title                string
	dst                  core.NavigationPath
	justAuthenticated    bool
	action               func(wnd core.Window)
	oneOfAuthorizedPerms []permission.ID
	onlyPublic           bool
	oneOfRoles           []role.ID
	submenu              *SubMenuBuilder
	dyn                  func(wnd core.Window, entry *MenuEntryBuilder)
}

func (b *MenuEntryBuilder) OneOf(perms ...permission.ID) *ScaffoldBuilder {
	b.oneOfAuthorizedPerms = append(b.oneOfAuthorizedPerms, perms...)
	return b.parent
}

func (b *MenuEntryBuilder) OneOfRole(roles ...role.ID) *ScaffoldBuilder {
	b.oneOfRoles = append(b.oneOfRoles, roles...)
	return b.parent
}

func (b *MenuEntryBuilder) Private() *ScaffoldBuilder {
	b.justAuthenticated = true
	return b.parent
}

// Public shows this entry for authenticated and non-authenticated users. See also [MenuEntryBuilder.OnlyPublic]
// and [MenuEntryBuilder.Private] and [MenuEntryBuilder.PublicOnly].
func (b *MenuEntryBuilder) Public() *ScaffoldBuilder {
	return b.parent
}

func (b *MenuEntryBuilder) PublicOnly() *ScaffoldBuilder {
	b.onlyPublic = true
	return b.parent
}

func (b *MenuEntryBuilder) Dynamic(fn func(wnd core.Window, entry *MenuEntryBuilder)) *ScaffoldBuilder {
	b.dyn = fn
	return b.parent
}

func (b *MenuEntryBuilder) Icon(icon core.SVG) *MenuEntryBuilder {
	b.icon = icon
	return b
}

func (b *MenuEntryBuilder) Custom(view core.View) *MenuEntryBuilder {
	b.customView = view
	return b
}

func (b *MenuEntryBuilder) Title(title string) *MenuEntryBuilder {
	b.title = title
	return b
}

func (b *MenuEntryBuilder) Forward(dst core.NavigationPath) *MenuEntryBuilder {
	b.dst = dst
	return b
}

func (b *MenuEntryBuilder) Action(fn func(wnd core.Window)) *MenuEntryBuilder {
	b.action = fn
	return b
}

type SubMenuBuilder struct {
	parent  *MenuEntryBuilder
	entries []*SubMenuEntryBuilder
}

func (b *SubMenuBuilder) Title(title string) *SubMenuBuilder {
	b.parent.title = title
	return b
}

func (b *SubMenuBuilder) Icon(icon core.SVG) *SubMenuBuilder {
	b.parent.icon = icon
	return b
}

func (b *SubMenuBuilder) MenuEntry() *SubMenuEntryBuilder {
	e := &SubMenuEntryBuilder{
		parent: b,
	}

	b.entries = append(b.entries, e)

	return e
}

func (b *SubMenuBuilder) OneOf(perms ...permission.ID) *ScaffoldBuilder {
	b.parent.oneOfAuthorizedPerms = append(b.parent.oneOfAuthorizedPerms, perms...)
	return b.parent.parent
}

func (b *SubMenuBuilder) OneOfRole(roles ...role.ID) *ScaffoldBuilder {
	b.parent.oneOfRoles = append(b.parent.oneOfRoles, roles...)
	return b.parent.parent
}

func (b *SubMenuBuilder) Private() *ScaffoldBuilder {
	b.parent.justAuthenticated = true
	return b.parent.parent
}

// Public shows this entry for authenticated and non-authenticated users. See also [MenuEntryBuilder.OnlyPublic]
// and [MenuEntryBuilder.Private] and [MenuEntryBuilder.PublicOnly].
func (b *SubMenuBuilder) Public() *ScaffoldBuilder {
	return b.parent.parent
}

func (b *SubMenuBuilder) PublicOnly() *ScaffoldBuilder {
	b.parent.onlyPublic = true
	return b.parent.parent
}

type SubMenuEntryBuilder struct {
	parent               *SubMenuBuilder
	title                string
	dst                  core.NavigationPath
	justAuthenticated    bool
	action               func(wnd core.Window)
	oneOfAuthorizedPerms []permission.ID
	onlyPublic           bool
	oneOfRoles           []role.ID
}

func (b *SubMenuEntryBuilder) Title(title string) *SubMenuEntryBuilder {
	b.title = title
	return b
}

func (b *SubMenuEntryBuilder) Forward(dst core.NavigationPath) *SubMenuEntryBuilder {
	b.dst = dst
	return b
}

func (b *SubMenuEntryBuilder) Action(fn func(wnd core.Window)) *SubMenuEntryBuilder {
	b.action = fn
	return b
}

func (b *SubMenuEntryBuilder) OneOf(perms ...permission.ID) *SubMenuBuilder {
	b.oneOfAuthorizedPerms = append(b.oneOfAuthorizedPerms, perms...)
	return b.parent
}

func (b *SubMenuEntryBuilder) OneOfRole(roles ...role.ID) *SubMenuBuilder {
	b.oneOfRoles = append(b.oneOfRoles, roles...)
	return b.parent
}

func (b *SubMenuEntryBuilder) Private() *SubMenuBuilder {
	b.justAuthenticated = true
	return b.parent
}

// Public shows this entry for authenticated and non-authenticated users. See also [MenuEntryBuilder.OnlyPublic]
// and [MenuEntryBuilder.Private] and [MenuEntryBuilder.PublicOnly].
func (b *SubMenuEntryBuilder) Public() *SubMenuBuilder {
	return b.parent
}

func (b *SubMenuEntryBuilder) PublicOnly() *SubMenuBuilder {
	b.onlyPublic = true
	return b.parent
}

type ScaffoldBuilder struct {
	cfg                     *Configurator
	alignment               ui.ScaffoldAlignment
	menu                    []*MenuEntryBuilder
	logoClick               func(wnd core.Window)
	logoImage               ui.DecoredView
	showLogin               bool
	breakpoint              *int
	footer                  core.View
	enableAutoFooter        bool
	footerBackgroundColor   ui.Color
	footerTextColor         ui.Color
	disableContentPaddingOn []core.NavigationPath
	height                  ui.Length
}

func (c *Configurator) NewScaffold() *ScaffoldBuilder {
	return &ScaffoldBuilder{
		cfg:       c,
		alignment: ui.ScaffoldAlignmentTop,

		showLogin:        true,
		enableAutoFooter: true,
		logoClick: func(wnd core.Window) {
			wnd.Navigation().ForwardTo(".", nil)
		},
	}
}

func (b *ScaffoldBuilder) Breakpoint(breakpoint int) *ScaffoldBuilder {
	b.breakpoint = &breakpoint
	return b
}

func (b *ScaffoldBuilder) Login(show bool) *ScaffoldBuilder {
	b.showLogin = show
	return b
}

func (b *ScaffoldBuilder) Alignment(alignment ui.ScaffoldAlignment) *ScaffoldBuilder {
	b.alignment = alignment
	return b
}

func (b *ScaffoldBuilder) Height(height ui.Length) *ScaffoldBuilder {
	b.height = height
	return b
}

// Logo sets an already allocated (image) component as the menubar image. Note, that you should match the height
// of 6rem or [ui.L96]. Also see [ui.TImage.EmbedAdaptive] to support light and dark mode switching automatically.
// See [ScaffoldBuilder.LogoAction] to declare a listener which receives [core.Window] for navigation.
func (b *ScaffoldBuilder) Logo(image ui.DecoredView) *ScaffoldBuilder {
	b.logoImage = image
	return b
}

func (b *ScaffoldBuilder) Footer(footer core.View) *ScaffoldBuilder {
	b.footer = footer
	b.enableAutoFooter = false
	return b
}

// FooterBackgroundColor is only applied for the automatically generated footer and not if a custom component
// is configured by user [ScaffoldBuilder.Footer].
func (b *ScaffoldBuilder) FooterBackgroundColor(color ui.Color) *ScaffoldBuilder {
	b.footerBackgroundColor = color
	return b
}

// FooterTextColor is only applied for the automatically generated footer and not if a custom component
// is configured by user [ScaffoldBuilder.Footer].
func (b *ScaffoldBuilder) FooterTextColor(color ui.Color) *ScaffoldBuilder {
	b.footerTextColor = color
	return b
}

// LogoAction allows a forward declaration using a function callback with the current window.
// By default, the [ScaffoldBuilder] applies an action to navigate forward to
// the index root view (which is .) but this behavior can be disabled by setting a nil function here.
func (b *ScaffoldBuilder) LogoAction(fn func(wnd core.Window)) *ScaffoldBuilder {
	b.logoClick = fn
	return b
}

// SubmenuEntry configures a new entry with a single level sub menu. The ora design guidelines do not
// specify more levels.
func (b *ScaffoldBuilder) SubmenuEntry(fn func(menu *SubMenuBuilder)) *ScaffoldBuilder {
	e := &MenuEntryBuilder{parent: b}
	b.menu = append(b.menu, e)

	if fn != nil {
		subMenu := &SubMenuBuilder{
			parent: e,
		}
		fn(subMenu)
		e.submenu = subMenu
	}

	return b
}

func (b *ScaffoldBuilder) MenuEntry() *MenuEntryBuilder {
	e := &MenuEntryBuilder{parent: b}
	b.menu = append(b.menu, e)
	return e
}

// NoFooterContentPadding disables the footers content padding on the given paths.
func (b *ScaffoldBuilder) NoFooterContentPadding(paths ...core.NavigationPath) *ScaffoldBuilder {
	b.disableContentPaddingOn = paths
	return b
}

func (b *ScaffoldBuilder) name() string {
	return b.cfg.applicationName
}

func (b *ScaffoldBuilder) registerLegalViews() {
	// 404 not found case
	b.cfg.RootView("_", func(wnd core.Window) core.View {
		// we introduce another indirection, so that we can use the iamSettings AFTER this builder completed
		return b.cfg.DecorateRootView(func(wnd core.Window) core.View {
			return alert.NotFound()
		})(wnd)
	})

}

// Decorator is a builder terminal and returns a decorator-like function.
func (b *ScaffoldBuilder) Decorator() func(wnd core.Window, view core.View) core.View {
	b.registerLegalViews()

	return func(wnd core.Window, view core.View) core.View {
		themeCfg := core.GlobalSettings[theme.Settings](wnd)

		var logo core.View
		if b.logoImage != nil {
			//logo = ui.HStack(b.logoImage).Frame(ui.Frame{}.Size("", "6rem"))
			logo = b.logoImage
		} else {
			logo = ui.Image().Adaptive(themeCfg.PageLogoLight, themeCfg.PageLogoDark).ObjectFit(ui.FitContain).Frame(ui.Frame{Height: ui.L80})
		}

		var menu []ui.ScaffoldMenuEntry

		for _, entry := range b.menu {
			if entry.dyn != nil {
				entry.dyn(wnd, entry)
			}

			if entry.justAuthenticated && !wnd.Subject().Valid() {
				continue
			}

			if wnd.Subject().Valid() && entry.onlyPublic {
				continue
			}

			if len(entry.oneOfAuthorizedPerms) > 0 {
				if !auth.OneOf(wnd.Subject(), entry.oneOfAuthorizedPerms...) {
					continue
				}
			}

			if len(entry.oneOfRoles) > 0 {
				hasRole := false
				for _, roleId := range entry.oneOfRoles {
					if wnd.Subject().HasRole(roleId) {
						hasRole = true
						break
					}
				}

				if !hasRole {
					continue
				}

			}

			icoSize := ui.L24
			if entry.title == "" {
				icoSize = ui.L40
			}

			eTitle := entry.title
			if strings.HasPrefix(eTitle, "@") {
				eTitle = wnd.Bundle().Resolve(eTitle)
			}

			sentry := ui.ScaffoldMenuEntry{
				Icon:  ui.If(entry.icon != nil, ui.Image().Embed(entry.icon).Frame(ui.Frame{}.Size(icoSize, icoSize))),
				Title: eTitle,
				Action: func() {
					if entry.action != nil {
						entry.action(wnd)
					}

					if entry.dst != "" {
						wnd.Navigation().ForwardTo(entry.dst, nil)
					}
				},
				MarkAsActiveAt: entry.dst,
			}

			if entry.customView != nil {
				sentry.Icon = entry.customView
			}

			if entry.submenu != nil {
				sentry.Action = nil
				for _, subentry := range entry.submenu.entries {
					// TODO this is a duplicate
					if subentry.justAuthenticated && !wnd.Subject().Valid() {
						continue
					}

					if wnd.Subject().Valid() && entry.onlyPublic {
						continue
					}

					if len(subentry.oneOfAuthorizedPerms) > 0 {
						if !auth.OneOf(wnd.Subject(), entry.oneOfAuthorizedPerms...) {
							continue
						}
					}

					if len(subentry.oneOfRoles) > 0 {
						hasRole := false
						for _, roleId := range subentry.oneOfRoles {
							if wnd.Subject().HasRole(roleId) {
								hasRole = true
								break
							}
						}

						if !hasRole {
							continue
						}

					}
					// TODO snap duplicate

					sentry.Menu = append(sentry.Menu, ui.ScaffoldMenuEntry{
						Title: subentry.title,
						Action: func() {
							if subentry.action != nil {
								subentry.action(wnd)
							}

							if subentry.dst != "" {
								wnd.Navigation().ForwardTo(subentry.dst, nil)
							}
						},
						MarkAsActiveAt: subentry.dst,
					})
				}
			}

			menu = append(menu, sentry)
		}

		menuDialogPresented := ScaffoldUserMenuPresentedState(wnd)

		if sessionManagement := b.cfg.sessionManagement; sessionManagement != nil && b.showLogin {
			if !wnd.Subject().Valid() {
				menu = append(menu, ui.ForwardScaffoldMenuEntry(wnd, flowbiteOutline.ArrowLeftToBracket, "Anmelden", sessionManagement.Pages.Login))
			} else {
				var avatarIcon core.View
				if id := wnd.Subject().Avatar(); id != "" {
					avatarIcon = avatar.URI(httpimage.URI(image.ID(id), image.FitCover, 40, 40))
				} else {
					avatarIcon = avatar.Text(wnd.Subject().Name()).Size(ui.L40)
				}

				menu = append(menu, ui.ScaffoldMenuEntry{
					Icon:  avatarIcon,
					Title: "",
					Action: func() {
						menuDialogPresented.Set(true)
					},
				})
			}
		}

		var schemeModeIcon core.SVG
		var schemeModeText string
		if wnd.Info().ColorScheme == core.Light {
			schemeModeIcon = flowbiteOutline.Moon
			schemeModeText = "Dunkle Darstellung verwenden"
		} else {
			schemeModeIcon = flowbiteOutline.Sun
			schemeModeText = "Helle Darstellung verwenden"
		}

		var scaffold = ui.Scaffold(b.alignment).Body(
			ui.VStack(
				ui.WindowTitle(b.name()),
				ui.IfFunc(b.cfg.sessionManagement != nil, func() core.View {
					return b.profileDialog(wnd, b.cfg.sessionManagement, menuDialogPresented)
				}),

				view,
				alert.BannerMessages(wnd),
				tracking.SupportRequestDialog(wnd), // be the last one, to guarantee to be on top
			).FullWidth(),
		).BottomView(lightDarkButton(schemeModeIcon, schemeModeText, func() {
			if wnd.Info().ColorScheme == core.Light {
				wnd.SetColorScheme(core.Dark)
			} else {
				wnd.SetColorScheme(core.Light)
			}
		})).
			Logo(ui.HStack(logo).Action(b.logoActionClick(wnd)).Frame(ui.Frame{Height: ui.L96})).
			Menu(menu...).Height(b.height)

		if b.breakpoint != nil {
			scaffold = scaffold.Breakpoint(*b.breakpoint)
		}

		noFooter := b.cfg.noFooter
		if len(noFooter) == 0 || (len(noFooter) > 0 && !slices.Contains(noFooter, wnd.Path())) {
			if b.footer != nil {
				scaffold = scaffold.Footer(b.footer)
			} else if b.enableAutoFooter {

				autoFooter := footer.Footer().
					Logo(ui.Image().Adaptive(themeCfg.PageLogoLight, themeCfg.PageLogoDark).ObjectFit(ui.FitContain).Frame(ui.Frame{Height: ui.L64})).
					Impress(themeCfg.Impress).
					PrivacyPolicy(themeCfg.PrivacyPolicy).
					TermsOfUse(themeCfg.TermsOfUse).
					ProviderName(themeCfg.ProviderName).
					TextColor(b.footerTextColor).
					BackgroundColor(b.footerBackgroundColor).
					Slogan(themeCfg.Slogan).
					GeneralTermsAndConditions(themeCfg.GeneralTermsAndConditions)

				if slices.Contains(b.disableContentPaddingOn, wnd.Path()) {
					autoFooter = autoFooter.ContentPadding("")
				}

				scaffold = scaffold.Footer(autoFooter)
			}
		}

		return scaffold
	}
}

func ScaffoldUserMenuPresentedState(wnd core.Window) *core.State[bool] {
	return core.StateOf[bool](wnd, "nago.scaffold.user.menu.presented")
}

func (b *ScaffoldBuilder) logoActionClick(wnd core.Window) func() {
	if b.logoClick == nil {
		return nil
	}

	return func() {
		b.logoClick(wnd)
	}
}

func (b *ScaffoldBuilder) hasAdminCenter(wnd core.Window) bool {
	if b.cfg.adminManagement == nil {
		return false
	}

	visibleEntries := b.cfg.adminManagement.QueryGroups(wnd.Subject(), "")
	return len(visibleEntries) > 0
}

func (b *ScaffoldBuilder) profileMenu(wnd core.Window) core.View {
	return uiuser.AccountView(wnd)
}

func (b *ScaffoldBuilder) profileDialog(wnd core.Window, sessionManagement *SessionManagement, state *core.State[bool]) core.View {
	if !state.Get() {
		return nil
	}

	isMobile := wnd.Info().SizeClass == core.SizeClassSmall

	var opts []alert.Option

	opts = append(opts, alert.Closeable())
	// TODO read how many impress, gtc etc entries we have and if we have all 4 use alert.Large
	if !isMobile && wnd.Info().Height > 600 {
		opts = append(opts, alert.Alignment(ui.TopTrailing), alert.ModalPadding(ui.Padding{}.All(ui.L80)))
	}

	return alert.Dialog("Nutzerkonto", b.profileMenu(wnd), state, opts...)
}

func lightDarkButton(icon core.SVG, text string, action func()) core.View {
	return ui.HStack(
		ui.ImageIcon(icon),
		ui.Text(text).Font(ui.Font{Weight: ui.HeadlineAndTitleFontWeight}),
	).Gap(ui.L8).
		HoveredBackgroundColor(ui.I1).
		Action(action).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L4)).
		Border(ui.Border{}.Radius(ui.L8))
}
