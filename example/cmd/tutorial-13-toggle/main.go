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

		cfg.RootView(".", func(wnd core.Window) core.View {
			checked := core.AutoState[bool](wnd)
			showAlert := core.AutoState[bool](wnd)

			return VStack(
				alert.Dialog("Achtung", Text(fmt.Sprintf("Deine Eingabe: %v", checked)), showAlert, alert.Ok()),
				Toggle(checked.Get()).InputChecked(checked),
				PrimaryButton(func() {
					showAlert.Set(true)
				}).Title("Check"),
			).Gap(L16).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
