package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/progress"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_50")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			pwdState := core.AutoState[string](wnd)
			indicator := user.CalculatePasswordStrength(pwdState.Get())
			return ui.VStack(
				ui.Text("hello world"),
				ui.TextField("Password", pwdState.Get()).InputValue(pwdState),
				progress.LinearProgress().Progress(indicator.ComplexityScale),
				uiuser.PasswordStrengthView(indicator),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())

		}))

	}).Run()
}
