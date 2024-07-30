package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/presentation/ui2/alert"
	"go.wdy.de/nago/web/vuejs"
	"strings"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			firstname := core.AutoState[string](wnd)
			showAlert := core.AutoState[bool](wnd)

			return ui.VStack(
				alert.Dialog("Achtung", fmt.Sprintf("Deine Eingabe: %v", firstname), showAlert, alert.Ok()),
				ui.TextField("hello world", firstname.Get()).InputValue(firstname),
				// you can re-use the state, but be careful of the effects
				ui.TextField("just numbers", numsOf(firstname.Get())).InputValue(firstname).Style(ora.TextFieldReduced),

				ui.PrimaryButton(func() {
					showAlert.Set(true)
				}).Title("Check"),
			).Gap(ora.L16).
				Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}

func numsOf(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}

	return sb.String()
}
