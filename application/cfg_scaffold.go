package application

import (
	_ "embed"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/glossary/docm"
	"go.wdy.de/nago/glossary/docm/markdown"
	"go.wdy.de/nago/glossary/docm/oraui"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/tracking"
)

//go:embed content_impress_de.md
var defaultImpress string

//go:embed content_gdpr_de.md
var defaultGDPR string

//go:embed content_policies_de.md
var defaultPolicies string

type MenuEntryBuilder struct {
	parent               *ScaffoldBuilder
	icon                 core.SVG
	title                string
	dst                  core.NavigationPath
	justAuthenticated    bool
	action               func(wnd core.Window)
	oneOfAuthorizedPerms []string
	onlyPublic           bool
}

func (b *MenuEntryBuilder) OneOf(perms ...string) *ScaffoldBuilder {
	b.oneOfAuthorizedPerms = append(b.oneOfAuthorizedPerms, perms...)
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

func (b *MenuEntryBuilder) Icon(icon core.SVG) *MenuEntryBuilder {
	b.icon = icon
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

type ScaffoldBuilder struct {
	cfg             *Configurator
	alignment       ui.ScaffoldAlignment
	lightLogo       core.SVG
	darkLogo        core.SVG
	policiesPath    core.NavigationPath
	policiesContent string
	gdprPath        core.NavigationPath
	gdprContent     string
	impressPath     core.NavigationPath
	impressContent  string
	menu            []MenuEntryBuilder
	logoClick       func()
}

func (c *Configurator) NewScaffold() *ScaffoldBuilder {
	return &ScaffoldBuilder{
		cfg:             c,
		alignment:       ui.ScaffoldAlignmentTop,
		policiesPath:    "legal/policies",
		policiesContent: defaultPolicies,
		impressPath:     "legal/impress",
		impressContent:  defaultImpress,
		gdprPath:        "legal/gdpr",
		gdprContent:     defaultGDPR,
	}
}

func (b *ScaffoldBuilder) LogoAction(fn func()) *ScaffoldBuilder {
	b.logoClick = fn
	return b
}

func (b *ScaffoldBuilder) Logo(svg core.SVG) *ScaffoldBuilder {
	b.lightLogo = svg
	b.darkLogo = svg
	return b
}

func (b *ScaffoldBuilder) LogoDark(svg core.SVG) *ScaffoldBuilder {
	b.darkLogo = svg
	return b
}

func (b *ScaffoldBuilder) LogoLight(svg core.SVG) *ScaffoldBuilder {
	b.lightLogo = svg
	return b
}

func (b *ScaffoldBuilder) MenuEntry() *MenuEntryBuilder {
	b.menu = append(b.menu, MenuEntryBuilder{parent: b})
	return &b.menu[len(b.menu)-1]
}

func (b *ScaffoldBuilder) hasIAM() bool {
	return b.cfg.iamSettings.Service != nil
}

func (b *ScaffoldBuilder) name() string {
	return b.cfg.applicationName
}

func (b *ScaffoldBuilder) registerLegalViews() {
	// 404 not found case
	b.cfg.RootView("_", func(wnd core.Window) core.View {
		// we introduce another indirection, so that we can use the iamSettings AFTER this builder completed
		return b.cfg.iamSettings.DecorateRootView(func(wnd core.Window) core.View {
			return ui.VStack(
				ui.WindowTitle("Nicht gefunden"),
				alert.Banner("Resource nicht gefunden", "Die Seite, Funktion oder Resource ist dauerhaft nicht verfügbar."),
			)
		})(wnd)
	})

	b.cfg.RootView(b.impressPath, func(wnd core.Window) core.View {
		// we introduce another indirection, so that we can use the iamSettings AFTER this builder completed
		return b.cfg.iamSettings.DecorateRootView(func(wnd core.Window) core.View {
			return oraui.Render(&docm.Document{Body: markdown.Parse(b.impressContent)})
		})(wnd)
	})

	b.cfg.RootView(b.policiesPath, func(wnd core.Window) core.View {
		// we introduce another indirection, so that we can use the iamSettings AFTER this builder completed
		return b.cfg.iamSettings.DecorateRootView(func(wnd core.Window) core.View {
			return oraui.Render(&docm.Document{Body: markdown.Parse(b.policiesContent)})
		})(wnd)
	})

	b.cfg.RootView(b.gdprPath, func(wnd core.Window) core.View {
		// we introduce another indirection, so that we can use the iamSettings AFTER this builder completed
		return b.cfg.iamSettings.DecorateRootView(func(wnd core.Window) core.View {
			return oraui.Render(&docm.Document{Body: markdown.Parse(b.gdprContent)})
		})(wnd)
	})
}

// Decorator is a builder terminal and returns a decorator-like function.
func (b *ScaffoldBuilder) Decorator() func(wnd core.Window, view core.View) core.View {
	b.registerLegalViews()

	return func(wnd core.Window, view core.View) core.View {
		var logo core.View
		if b.lightLogo != nil || b.darkLogo != nil {
			logo = ui.Image().EmbedAdaptive(b.lightLogo, b.darkLogo).Frame(ui.Frame{}.Size(ui.L160, ui.L160))
		}

		var menu []ui.ScaffoldMenuEntry

		for _, entry := range b.menu {
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

			icoSize := ui.L24
			if entry.title == "" {
				icoSize = ui.L40
			}

			menu = append(menu, ui.ScaffoldMenuEntry{
				Icon:  ui.If(entry.icon != nil, ui.Image().Embed(entry.icon).Frame(ui.Frame{}.Size(icoSize, icoSize))),
				Title: entry.title,
				Action: func() {
					if entry.action != nil {
						entry.action(wnd)
					}

					if entry.dst != "" {
						wnd.Navigation().ForwardTo(entry.dst, nil)
					}
				},
				MarkAsActiveAt: entry.dst,
			})
		}

		menuDialogPresented := core.AutoState[bool](wnd)
		iamCfg := b.cfg.iamSettings

		if b.hasIAM() {
			if !wnd.Subject().Valid() {
				menu = append(menu, ui.ForwardScaffoldMenuEntry(wnd, heroSolid.ArrowLeftEndOnRectangle, "Anmelden", iamCfg.Login.ID))
			} else {
				menu = append(menu, ui.ScaffoldMenuEntry{
					Icon:  avatar.Text(wnd.Subject().Name()),
					Title: "",
					Action: func() {
						menuDialogPresented.Set(true)
					},
				})
			}
		}

		return ui.Scaffold(b.alignment).Body(
			ui.VStack(
				ui.WindowTitle(b.name()),
				b.profileDialog(wnd, menuDialogPresented),

				view,
				alert.BannerMessages(wnd),
				tracking.SupportRequestDialog(wnd), // be the last one, to guarantee to be on top
			).FullWidth(),
		).Logo(ui.HStack(logo).Action(b.logoClick)).
			Menu(menu...)
	}
}

func (b *ScaffoldBuilder) isAdmin(wnd core.Window) bool {
	return auth.OneOf(wnd.Subject(), iam.ReadGroup, iam.ReadPermission, iam.ReadRole, iam.ReadUser)
}

func (b *ScaffoldBuilder) profileMenu(wnd core.Window) core.View {
	return ui.VStack(
		ui.HStack(
			avatar.Text(wnd.Subject().Name()).Size(ui.L120),
			ui.VStack(
				ui.Text(wnd.Subject().Name()).Font(ui.Title),
				ui.Text(wnd.Subject().Email()),
				ui.HStack(
					colorSchemeToggle(wnd),
					ui.If(b.isAdmin(wnd), ui.SecondaryButton(func() {
						wnd.Navigation().ForwardTo(b.cfg.iamSettings.Dashboard.ID, nil)
					}).PreIcon(heroOutline.UserGroup).AccessibilityLabel("Nutzer verwalten")),
					ui.SecondaryButton(func() {
						service := b.cfg.iamSettings.Service
						service.Logout(wnd.SessionID())
						wnd.UpdateSubject(service.Subject(wnd.SessionID()))
						wnd.Navigation().ForwardTo(b.cfg.iamSettings.Logout.ID, nil)
					}).PreIcon(heroOutline.ArrowLeftStartOnRectangle).AccessibilityLabel("Abmelden"),
				).FullWidth().Gap(ui.L8).Alignment(ui.Leading),
			).Gap(ui.L4).Alignment(ui.Leading),
		).Gap(ui.L16),
		ui.HLineWithColor(ui.ColorAccent),
		ui.HStack(
			ui.SecondaryButton(func() {

			}).Title("Konto verwalten"),
		).FullWidth().Alignment(ui.Trailing),
		ui.HStack(
			ui.Text("Datenschutzerklärung").Font(ui.Small).Action(func() {
				wnd.Navigation().ForwardTo(b.gdprPath, nil)
			}),
			ui.Text("Nutzungsbedingungen").Font(ui.Small).Action(func() {
				wnd.Navigation().ForwardTo(b.policiesPath, nil)
			}),
			ui.Text("Impressum").Font(ui.Small).Action(func() {
				wnd.Navigation().ForwardTo(b.impressPath, nil)
			}),
		).FullWidth().Gap(ui.L8).Padding(ui.Padding{Top: ui.L16}),
	).Alignment(ui.Leading).FullWidth()
}

func (b *ScaffoldBuilder) profileDialog(wnd core.Window, state *core.State[bool]) core.View {
	if !state.Get() {
		return nil
	}

	subj := wnd.Subject()
	return alert.Dialog(subj.Name(), b.profileMenu(wnd), state, alert.Closeable(), alert.Alignment(ui.TopTrailing), alert.ModalPadding(ui.Padding{}.All(ui.L80)))
}

func colorSchemeToggle(wnd core.Window) core.View {
	var icon core.SVG
	if wnd.Info().ColorScheme == core.Dark {
		icon = heroOutline.Sun
	} else {
		icon = heroOutline.Moon
	}

	return ui.SecondaryButton(func() {
		current := wnd.Info().ColorScheme
		if current == core.Dark {
			current = core.Light
		} else {
			current = core.Dark
		}

		wnd.SetColorScheme(current)
	}).AccessibilityLabel("Modus Hell und Dunkel umschalten").PreIcon(icon)
}
