package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

type MyCustomColors struct {
	MySuperColor ora.Color
}

func (m MyCustomColors) Default(scheme ora.ColorScheme) ora.ColorSet {
	if scheme == ora.Light {
		return MyCustomColors{MySuperColor: "#cd29ff"}
	}

	return MyCustomColors{MySuperColor: "#12ffc8"}
}

func (m MyCustomColors) Namespace() ora.NamespaceName {
	return "myCustomColor"
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.ColorSet(ora.Light, MyCustomColors{
			MySuperColor: "#ff0000",
		})

		cfg.Component(".", func(wnd core.Window) core.View {
			colors := core.ColorSet[MyCustomColors](wnd)
			defCols := core.ColorSet[ora.Colors](wnd)
			return ui.VStack(
				ui.Text("hello world"),

				ui.Modal(
					ui.Box(ui.BoxLayout{Center: ui.VStack(
						ui.Text("ich bin modal2"),
						ui.VStack(ui.Text("hover me")).HoveredBackgroundColor("#ff0000").FocusedBackgroundColor(),
					).BackgroundColor(defCols.M1).Frame(ora.Frame{}.Size("300dp", "200dp"))}).
						BackgroundColor(colors.MySuperColor),
				),
			)
		})
	}).Run()
}
