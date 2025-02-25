package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "8fb8724f-e604-444c-9671-58d07dd76164"))

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			return ui.VStack(
				ui.PrimaryButton(func() {
					if err := wnd.Clipboard().SetText(fmt.Sprintf("hello world: %v", time.Now())); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
				}).Title("copy to clipboard"),
			).Frame(ui.Frame{}.MatchScreen())

		}))

	}).Run()
}
