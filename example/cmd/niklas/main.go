package main

import (
	_ "embed"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/example/cmd/niklas/pages"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Index("example")
		cfg.Name("example-niklas")
		cfg.Serve(vuejs.Dist())

		cfg.Page("morty", func(wire ui.Wire) *ui.Page {
			return pages.Morty(wire)
		})

		cfg.Page("example", func(wire ui.Wire) *ui.Page {
			return pages.Example(wire)
		})

	}).Run()
}
