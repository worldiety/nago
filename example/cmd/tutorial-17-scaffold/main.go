package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	icons2 "go.wdy.de/nago/presentation/icons/flowbite/outline"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {

			return ui.Scaffold(ora.ScaffoldAlignmentTop).
				Body(ui.Text("hello body")).
				Logo(ui.Image().Embed(icons.SpeakerWave).Frame(ora.Frame{}.Size(ora.L20, ora.L20))).
				Menu(
					ui.MenuEntry{
						Icon:       nil,
						IconActive: nil,
						Title:      "Entry 1",
						Action: func() {
							fmt.Println("clicked entry 1")
						},
						Factory:  "",
						Menu:     nil,
						Badge:    "",
						Expanded: false,
					},
					ui.MenuEntry{
						Icon:       ui.Image().Embed(icons.User).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						IconActive: nil,
						Title:      "Entry 2",
						Action:     nil,
						Factory:    ".",
						Menu: []ui.MenuEntry{
							{
								Title: "sub a",
							},
							{
								Title: "sub b",
							},
						},
						Badge:    "",
						Expanded: false,
					},

					ui.MenuEntry{
						Icon:       ui.Image().Embed(icons.Window).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						IconActive: ui.Image().Embed(icons2.Grid).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						Title:      "Entry 3",
						Action:     nil,
						Factory:    ".",
						Menu:       nil,
						Badge:      "42",
						Expanded:   false,
					},
					ui.MenuEntry{
						Icon:       ui.Image().Embed(icons.User).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						IconActive: ui.Image().Embed(icons2.User).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
						Title:      "nested top",
						Action:     nil,
						Factory:    ".",
						Menu: []ui.MenuEntry{
							{
								Title: "sub 1",
							},
							{
								Title: "sub 2",
							},
						},
						Badge:    "",
						Expanded: false,
					},
				)
		})
	}).Run()
}
