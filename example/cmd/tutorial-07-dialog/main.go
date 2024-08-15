package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			isPresented := core.AutoState[bool](wnd)
			return VStack(
				Text("show dialog").Action(func() {
					isPresented.Set(true)
				}),
				If(isPresented.Get(), Modal(
					Dialog(Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.")).
						Title(Text("Titel")).
						Footer(PrimaryButton(func() {
							isPresented.Set(false)
						}).Title("Schließen")),
				)),
			).Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
