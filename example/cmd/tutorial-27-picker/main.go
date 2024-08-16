package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
	"go.wdy.de/nago/web/vuejs"
)

type Person struct {
	Vorname, Nachname string
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		frodo := Person{"Frodo", "Beutlin"}
		personen := []Person{
			{"Bilbo", "Beutlin"},
			frodo,
			{"Peppin", "Tuk"},
		}

		cfg.RootView(".", func(wnd core.Window) core.View {
			personState := core.AutoState[[]Person](wnd).From(func() []Person {
				return []Person{frodo}
			})
			return VStack(
				picker.Picker[Person]("Hobbit", personen, personState).
					SupportingText("Wähle jemanden aus").
					Title("Hobbitse").
					MultiSelect().
					//ErrorText("Falsch").
					Frame(Frame{Width: L320}),
				PrimaryButton(func() {
					fmt.Println(personState)
				}).Title("print selected"),
			).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
