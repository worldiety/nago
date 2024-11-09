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
	"time"
)

//go:embed content_impress_de.md
var defaultImpress string

//go:embed content_gdpr_de.md
var defaultGDPR string

//go:embed content_policies_de.md
var defaultPolicies string

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

func (b *ScaffoldBuilder) Logo(svg core.SVG) *ScaffoldBuilder {
	b.lightLogo = svg
	b.darkLogo = svg
	return b
}

func (b *ScaffoldBuilder) DarkLogo(svg core.SVG) *ScaffoldBuilder {
	b.darkLogo = svg
	return b
}

func (b *ScaffoldBuilder) LightLogo(svg core.SVG) *ScaffoldBuilder {
	b.lightLogo = svg
	return b
}

func (b *ScaffoldBuilder) hasIAM() bool {
	return b.cfg.iamSettings.Service != nil
}

func (b *ScaffoldBuilder) name() string {
	return b.cfg.applicationName
}

func (b *ScaffoldBuilder) registerLegalViews() {
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
		if wnd.Info().ColorScheme == core.Dark && !b.darkLogo.Empty() {
			logo = ui.Image().Embed(b.darkLogo).Frame(ui.Frame{}.Size(ui.L160, ui.L160))
		} else if !b.lightLogo.Empty() {
			logo = ui.Image().Embed(b.lightLogo).Frame(ui.Frame{}.Size(ui.L160, ui.L160))
		}

		menuDialogPresented := core.AutoState[bool](wnd)
		iamCfg := b.cfg.iamSettings
		var menu []ui.ScaffoldMenuEntry
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

		menu = append(menu, ui.ScaffoldMenuEntry{
			Title: "Test snack",
			Action: func() {
				alert.ShowMessage(wnd, alert.Message{"snack it", "nom nom" + time.Now().String()})
			},
		})

		return ui.Scaffold(b.alignment).Body(
			ui.VStack(
				ui.WindowTitle(b.name()),
				b.profileDialog(wnd, menuDialogPresented),

				view,
				alert.MessageList(wnd),
				tracking.SupportRequestDialog(wnd), // be the last one, to guarantee to be on top
			).FullWidth(),
		).Logo(logo).
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
