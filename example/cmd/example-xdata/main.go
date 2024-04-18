package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nago.demo.xdata")
		personRepo := application.SloppyRepository[Person, PersonID](cfg)
		if err := initUsers(personRepo); err != nil {
			panic(err)
		}

		cfg.Serve(vuejs.Dist())

		persons := NewPersonService(personRepo)
		cfg.Component("hello", func(wnd core.Window) core.Component {
			return dataPage(wnd, persons)
		})

		cfg.Component("button", func(wnd core.Window) core.Component {
			return ui.NewButton(func(btn *ui.Button) {
				btn.Caption().Set("hello world")
				btn.Action().Set(func() {
					fmt.Println("clicked btn")
				})
			})
		})

	}).Run()
}
