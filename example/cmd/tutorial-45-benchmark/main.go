package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			var tmp []core.View

			tmp = append(tmp,
				ui.Text(fmt.Sprintf("%v", time.Now())),
				ui.PrimaryButton(nil).Title("hello"),
				ui.PrimaryButton(nil).Title("world"),
				ui.PrimaryButton(nil).Title("see"),
				ui.PrimaryButton(nil).Title("how"),
				ui.PrimaryButton(nil).Title("slow"),
				ui.PrimaryButton(nil).Title("the hover"),
			)

			for i := range 1000 {
				// just make it a bit more complex
				tmp = append(tmp, ui.VStack(
					ui.HStack(
						ui.Text(fmt.Sprintf("Zeile %d", i)),
					),
				))
			}

			return ui.RedrawAtFixedRate(wnd, time.Second, ui.VStack(tmp...).FullWidth())
		}))

	}).Run()
}
