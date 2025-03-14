package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_53")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Login(false).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			done := core.AutoState[bool](wnd)
			fmt.Println("render was called")

			return ui.VStack(
				ui.Text("hello world!"),
				ui.If(done.Get(), ui.Text("timer is done")),
				ui.CountDown(time.Second*10).Action(func() {
					done.Set(true)
				}).Frame(ui.Frame{}.FullWidth()),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())

		}))

	}).Run()
}
