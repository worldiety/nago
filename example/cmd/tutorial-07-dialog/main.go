package main

import (
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

		cfg.Component(".", func(wnd core.Window) core.View {
			isPresented := core.AutoState[bool](wnd)
			return ui.VStack(
				ui.Text("show dialog").Action(func() {
					isPresented.Set(true)
				}),
				ui.If(isPresented.Get(), ui.Modal(
					ui.Dialog(ui.Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.")).
						Title(ui.Text("Titel")).
						Footer(ui.Text("Schließen").Action(func() {
							isPresented.Set(false)
						})),
				)),
			).Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}
