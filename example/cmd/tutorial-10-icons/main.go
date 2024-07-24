package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/maps"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	outlineIcons "go.wdy.de/nago/presentation/icons/hero/outline"
	solidIcons "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(
				slices.Collect(func(yield func(view core.View) bool) {
					yield(ui.Text("trigger invalidate").
						Action(func() {
							fmt.Println("causing a render roundtrip, to see if draw-by-reference cache kicks in")
						}))

					for _, key := range maps.SortedKeys(solidIcons.All) {
						yield(ui.HStack(
							ui.Text(key),
							ui.Image().Embed(solidIcons.All[key]).
								FillColor("#ff0000").
								Padding(ora.Padding{}.All(ora.L4)).
								Border(ora.Border{}.Circle().Color("#00ff00").Width(ora.L4)).
								Frame(ora.Frame{}.Size(ora.L44, ora.L44)),
							ui.Image().Embed(outlineIcons.All[key]).
								StrokeColor("#0000ff").
								Padding(ora.Padding{}.All(ora.L4)).
								Border(ora.Border{}.Circle().Color("#00ff00").Width(ora.L4)).
								Frame(ora.Frame{}.Size(ora.L44, ora.L44)),
						))

					}

				})...,
			).Alignment(ora.Trailing)
		})
	}).Run()
}
