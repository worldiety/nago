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

		cfg.Component(".", func(wnd core.Window) core.View {
			colors := core.ColorSet[ora.Colors](wnd)
			return ui.VStack(
				ui.Text("hello world"),
				ui.Modal(
					ui.Box(ui.BoxLayout{Center: ui.VStack(
						ui.Text("ich bin modal"),
						ui.VStack(ui.Text("hover me")).Action(func() {
							fmt.Println("clicked")
						}).
							HoveredBackgroundColor("#ff0000").
							PressedBackgroundColor("#0000ff").
							FocusedBackgroundColor("#00ffff").
							BackgroundColor("#00ff00"),
					).BackgroundColor(colors.P1).Frame(ora.Frame{}.Size("300dp", "200dp"))}).BackgroundColor(colors.P0.WithBrightness(30).WithTransparency(60)),
				),
			) //
		})
	}).Run()
}
