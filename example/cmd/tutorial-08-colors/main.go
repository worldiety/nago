package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

// MyCustomColors is a ColorSet which provides a namespace and the type safe color fields.
type MyCustomColors struct {
	// Colors must be flat in this struct, public and of type color
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
			oraColors := core.ColorSet[ora.Colors](wnd)
			return ui.VStack(
				ui.Text("hello world").Color(oraColors.I0).BackgroundColor(oraColors.M0),
				ui.FilledButton(colors.MySuperColor, nil).Title("my super button"),
			).Gap(ora.L16).Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}
