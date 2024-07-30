package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/presentation/ui2/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			checked := core.AutoState[bool](wnd)
			showAlert := core.AutoState[bool](wnd)

			return ui.VStack(
				alert.Dialog("Achtung", fmt.Sprintf("Deine Eingabe: %v", checked), showAlert, alert.Ok()),
				ui.Checkbox(checked.Get()).InputChecked(checked),
				ui.HStack(
					ui.Checkbox(checked.Get()).InputChecked(checked),
					ui.Text("check right").Action(func() {
						checked.Set(!checked.Get())
					}),
				),
				ui.HStack(
					ui.Text("check left").Action(func() {
						checked.Set(!checked.Get())
					}),
					ui.Checkbox(checked.Get()).InputChecked(checked),
				).Gap(ora.L16),
				ui.PrimaryButton(func() {
					showAlert.Set(true)
				}).Title("Check"),
			).Gap(ora.L16).
				Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}
