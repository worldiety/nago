package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example 2")

		persons := application.SloppyRepository[Person, PersonID](cfg)
		if err := initUsers(persons); err != nil {
			panic(err)
		}

		cfg.Serve(vuejs.Dist())

		cfg.Page("hello", func(wire ui.Wire) *ui.Page {
			return dataPage(wire, persons)
		})
	}).Run()
}
