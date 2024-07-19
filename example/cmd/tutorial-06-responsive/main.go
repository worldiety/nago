package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Resource(application.StaticBytes{})

		cfg.Component(".", func(wnd core.Window) core.View {
			firstname := core.AutoState[string](wnd)
			lastname := core.AutoState[string](wnd)

			return ui.VStack(
				ui.ViewThatMatches(wnd,
					ui.SizeClass(ora.SizeClass2XL, ui.HStack(
						ui.TextField("Vorname", firstname),
						ui.TextField("Nachname", lastname),
					)),
					ui.SizeClass(ora.SizeClassSmall, ui.VStack(
						ui.TextField("Vorname", firstname),
						ui.TextField("Nachname", lastname),
					)),
					ui.SizeClass(ora.SizeClassMedium, ui.Text("so mittel")),
				),
				ui.Text(fmt.Sprintf("hello world %s", wnd.Info().SizeClass)),
			).
				Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}
