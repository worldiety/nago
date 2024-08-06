package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			checked := core.AutoState[bool](wnd)
			showAlert := core.AutoState[bool](wnd)

			return VStack(
				alert.Dialog("Achtung", fmt.Sprintf("Deine Eingabe: %v", checked), showAlert, alert.Ok()),
				Checkbox(checked.Get()).InputChecked(checked),
				HStack(
					Checkbox(checked.Get()).InputChecked(checked),
					Text("check right").Action(func() {
						checked.Set(!checked.Get())
					}),
				),
				HStack(
					Text("check left").Action(func() {
						checked.Set(!checked.Get())
					}),
					Checkbox(checked.Get()).InputChecked(checked),
				).Gap(L16),
				PrimaryButton(func() {
					showAlert.Set(true)
				}).Title("Check"),
			).Gap(L16).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
