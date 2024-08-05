package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui2"
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
			secret := core.AutoState[string](wnd)
			showAlert := core.AutoState[bool](wnd)

			return VStack(
				alert.Dialog("Achtung", Text(fmt.Sprintf("Deine Eingabe: %v\nsecret: %v", firstname, secret)), showAlert, alert.Ok()),
				TextField("hello world", firstname.Get()).InputValue(firstname),
				// you can re-use the state, but be careful of the effects
				TextField("just numbers", numsOf(firstname.Get())).
					InputValue(firstname).
					KeyboardType(KeyboardInteger).
					Style(TextFieldReduced),

				TextField("text area", "hello\nworld").Lines(3),
				PrimaryButton(func() {
					showAlert.Set(true)
				}).Title("Check"),

				PasswordField("your secret").InputValue(secret),
			).Gap(L16).
				Frame(Frame{}.MatchScreen())
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
