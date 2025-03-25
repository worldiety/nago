package main

import (
	"go.wdy.de/nago/application"
	cfgrest "go.wdy.de/nago/application/rest/cfg"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/swagger"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_56")
		cfg.Serve(vuejs.Dist())
		cfg.Serve(swagger.Dist())

		std.Must(cfgrest.Enable(cfg))

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(Text("hello world")).
				Frame(Frame{}.MatchScreen())

		})
	}).
		Run()
}
