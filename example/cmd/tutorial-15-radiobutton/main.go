package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/presentation/ui2/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			stateGroup := AutoRadioStateGroup(wnd, 3)
			if stateGroup.SelectedIndex() == -1 {
				stateGroup.SetSelectedIndex(1)
			}
			showAlert := core.AutoState[bool](wnd)

			return VStack(
				alert.Dialog("Achtung", fmt.Sprintf("Deine Eingabe: %v", stateGroup.SelectedIndex()), showAlert, alert.Ok()),
				VStack(Each2(stateGroup.States, func(idx int, checked *core.State[bool]) core.View {
					return HStack(
						RadioButton(checked.Get()).
							InputChecked(checked),
						Text(fmt.Sprintf("Option %d", idx)).
							Action(func() {
								stateGroup.SetSelectedIndex(idx)
							}),
					)
				})...),

				PrimaryButton(func() {
					showAlert.Set(true)
				}).Title("Check"),
			).Gap(L16).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
