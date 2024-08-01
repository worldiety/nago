package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	icons2 "go.wdy.de/nago/presentation/icons/flowbite/outline"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed ORA_logo.svg
var appIco application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		// update the global app icon
		cfg.AppIcon(cfg.Resource(appIco))

		cfg.Component(".", func(wnd core.Window) core.View {

			return ui.Scaffold(ora.ScaffoldAlignmentTop).
				Body(ui.VStack(
					// this only causes the side effect of setting the current page title
					ui.WindowTitle("scaffold example"),
					ui.Text("hello body"),
				)).
				Logo(ui.Image().Embed(icons.SpeakerWave).Frame(ora.Frame{}.Size(ora.L20, ora.L20))).
				Menu(
					ui.ScaffoldMenuEntry{
						Title: "Entry 1",
						Action: func() {
							fmt.Println("clicked entry 1")
						},
					},
					ui.ScaffoldMenuEntry{
						Icon:           ui.Image().Embed(icons.User).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						Title:          "Entry 2",
						MarkAsActiveAt: ".",
						Menu: []ui.ScaffoldMenuEntry{
							{
								Title: "sub a",
							},
							{
								Title: "sub b",
							},
						},
					},

					ui.ScaffoldMenuEntry{
						Icon:           ui.Image().Embed(icons.Window).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						IconActive:     ui.Image().Embed(icons2.Grid).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						Title:          "Entry 3",
						MarkAsActiveAt: ".",
					},
					ui.ScaffoldMenuEntry{
						Icon:           ui.Image().Embed(icons.User).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						IconActive:     ui.Image().Embed(icons2.User).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						Title:          "nested top",
						MarkAsActiveAt: ".",
						Menu: []ui.ScaffoldMenuEntry{
							{
								Title: "sub 1",
							},
							{
								Title: "sub 2",
							},
						},
					},
				)
		})
	}).Run()
}
