package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			redraws := core.AutoState[int](wnd)
			redraws.Set(redraws.Get() + 1)

			return ui.RedrawAtFixedRate(wnd, time.Second, ui.Text(fmt.Sprintf("redraw: %v", redraws)))
		})
	}).Run()
}
