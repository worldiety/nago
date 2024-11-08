package application

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/tracking"
)

type ScaffoldBuilder struct {
	cfg       *Configurator
	alignment ui.ScaffoldAlignment
	lightLogo core.SVG
	darkLogo  core.SVG
}

func (c *Configurator) NewScaffold() *ScaffoldBuilder {
	return &ScaffoldBuilder{
		cfg:       c,
		alignment: ui.ScaffoldAlignmentTop,
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

// Decorator is a builder terminal and returns a decorator-like function.
func (b *ScaffoldBuilder) Decorator() func(wnd core.Window, view core.View) core.View {
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

		return ui.Scaffold(b.alignment).Body(
			ui.VStack(
				ui.WindowTitle(b.name()),
				b.profileDialog(wnd, menuDialogPresented),

				view,
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
			ui.Text("Datenschutzerklärung").Font(ui.Small),
			ui.Text("Nutzungsbedingungen").Font(ui.Small),
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
