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

				ui.PrimaryButton(func() {
					fmt.Println("clicked the real primary")
				}).Title("primary button"),

				ui.FilledButton(colors.I0, func() {
					fmt.Println("clicked the fake primary")
				}).Title("primary button"),

				ui.Secondary(func() {

				}).Title("secondary button"),

				ui.Tertiary(func() {

				}).Title("tertiary button"),

				ui.FilledButton("#EF8A97", func() {
					fmt.Println("clicked the fake primary")
				}).TextColor("#ffffff").Title("other color"),
			).Gap(ora.L16).Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}
