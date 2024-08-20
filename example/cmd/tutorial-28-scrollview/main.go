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
			return VStack(
				Text("vertical scroll view"),
				ScrollView(
					VStack(
						Image().URI("https://picsum.photos/id/12/200/300").Frame(Frame{Width: "200dp", Height: "300dp"}),
						Image().URI("https://picsum.photos/id/13/200/300").Frame(Frame{Width: "200dp", Height: "300dp"}),
						Image().URI("https://picsum.photos/id/14/200/300").Frame(Frame{Width: "200dp", Height: "300dp"}),
					),
				).Frame(Frame{Width: "200dp", Height: "450dp"}),

				Text("horizontal scroll view"),
				ScrollView(
					HStack(
						Image().URI("https://picsum.photos/id/12/200/300").Frame(Frame{Width: "200dp", Height: "300dp"}),
						Image().URI("https://picsum.photos/id/13/200/300").Frame(Frame{Width: "200dp", Height: "300dp"}),
						Image().URI("https://picsum.photos/id/14/200/300").Frame(Frame{Width: "200dp", Height: "300dp"}),
					),
				).Axis(ScrollViewAxisHorizontal).
					Frame(Frame{Width: "300dp", Height: "300dp"}),
			).Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
