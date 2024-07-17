package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Grid(
					ui.GridCell(ui.Text("1").BackgroundColor("#ff0000").Frame(ora.Frame{}.Size(ora.L44, ora.L20))),
					ui.GridCell(ui.Text("2").BackgroundColor("#00ff00").Frame(ora.Frame{}.Size(ora.L20, ora.L44))),
					ui.GridCell(ui.Text("3").BackgroundColor("#0000ff").Frame(ora.Frame{}.Size(ora.L44, ora.L44))),
					ui.GridCell(ui.Text("4").BackgroundColor("#ffff00").Frame(ora.Frame{}.Size(ora.L44, ora.L44))),
					ui.GridCell(ui.Text("5").BackgroundColor("#ff00ff").Frame(ora.Frame{}.Size(ora.L120, ora.L44))),
					ui.GridCell(ui.Text("6").BackgroundColor("#00ffff").Frame(ora.Frame{}.Size(ora.L20, ora.L20))),
				).Rows(2),
			).
				Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}
