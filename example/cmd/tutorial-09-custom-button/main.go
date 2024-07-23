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
			return ui.VStack(
				MyButton("First button", func() {
					fmt.Println("clicked 1")
				}),
				MyButton("Second button", func() {
					fmt.Println("clicked 2")
				}),
			).Gap(ora.L16).BackgroundColor("#4C4B5F").Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}

func MyButton(caption string, action func()) core.View {
	return ui.VStack(ui.Text(caption).Color("#ffffff")).Action(action).
		HoveredBackgroundColor(ora.Color("#EF8A97").WithTransparency(10)).
		PressedBackgroundColor(ora.Color("#EF8A97").WithTransparency(25)).
		PressedBorder(ora.Border{}.Circle().Color("#ffffff").Width(ora.L2)).
		FocusedBorder(ora.Border{}.Circle().Color("#ffffff").Width(ora.L2)).
		BackgroundColor("#EF8A97").
		Padding(ora.Padding{}.Horizontal(ora.L16).Vertical(ora.L8)).
		Border(ora.Border{}.Circle().Color("#00000000").Width(ora.L2)) // add invisible default border, to avoid dimension changes
}
