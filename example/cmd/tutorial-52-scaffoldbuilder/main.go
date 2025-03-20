package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_52")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Login(false).
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			MenuEntry().Title("hello").Forward(".").Public().
			SubmenuEntry(func(menu *application.SubMenuBuilder) {
				menu.Title("sub menu")
				menu.MenuEntry().Title("first").Action(func(wnd core.Window) {
					fmt.Println("clicked the first entry")
				})
				menu.MenuEntry().Title("second").Forward(".").Public()
			}).
			Breakpoint(1000).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			return ui.VStack(
				ui.Text("hello world!"),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())

		}))

	}).Run()
}
