package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Image().Embed(icons.SpeakerWave),
				ui.Text("hello world")).
				Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}
