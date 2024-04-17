package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example 2")

		personRepo := application.SloppyRepository[Person, PersonID](cfg)
		if err := initUsers(personRepo); err != nil {
			panic(err)
		}

		cfg.Serve(vuejs.Dist())

		persons := NewPersonService(personRepo)
		cfg.Page("hello", func(wire ui.Wire) *ui.Page {
			return dataPage(wire, persons)
		})

		cfg.Component("button", func(wire ui.Wire) ui.LiveComponent {
			return ui.NewButton(func(btn *ui.Button) {
				btn.Caption().Set("hello world")
				btn.Action().Set(func() {
					fmt.Println("clicked btn")
				})
			})
		})

	}).Run()
}
