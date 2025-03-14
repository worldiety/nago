package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_54")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Login(false).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			src := core.AutoState[string](wnd).Init(func() string {
				return `# hello h1

* some
* bullet
* point

in my _markdown_!
`
			}).Observe(func(newValue string) {
				fmt.Println("got new value:", newValue)
			})

			return ui.VStack(
				ui.Text("hello world!"),
				ui.CodeEditor(src.Get()).
					InputValue(src).
					Frame(ui.Frame{Height: ui.L560}.FullWidth()).
					Language("markdown"),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())

		}))

	}).Run()
}
