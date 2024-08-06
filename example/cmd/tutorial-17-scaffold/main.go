package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	icons2 "go.wdy.de/nago/presentation/icons/flowbite/outline"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
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

			return Scaffold(ScaffoldAlignmentTop).
				Body(VStack(
					// this only causes the side effect of setting the current page title
					WindowTitle("scaffold example"),
					Text("hello body"),
				)).
				Logo(Image().Embed(icons.SpeakerWave).Frame(Frame{}.Size(L20, L20))).
				Menu(
					ScaffoldMenuEntry{
						Title: "Entry 1",
						Action: func() {
							fmt.Println("clicked entry 1")
						},
					},
					ScaffoldMenuEntry{
						Icon:           Image().Embed(icons.User).Frame(Frame{}.Size(L20, L20)),
						Title:          "Entry 2",
						MarkAsActiveAt: ".",
						Menu: []ScaffoldMenuEntry{
							{
								Title: "sub a",
							},
							{
								Title: "sub b",
							},
						},
					},

					ScaffoldMenuEntry{
						Icon:           Image().Embed(icons.Window).Frame(Frame{}.Size(L20, L20)),
						IconActive:     Image().Embed(icons2.Grid).Frame(Frame{}.Size(L20, L20)),
						Title:          "Entry 3",
						MarkAsActiveAt: ".",
					},
					ScaffoldMenuEntry{
						Icon:           Image().Embed(icons.User).Frame(Frame{}.Size(L20, L20)),
						IconActive:     Image().Embed(icons2.User).Frame(Frame{}.Size(L20, L20)),
						Title:          "nested top",
						MarkAsActiveAt: ".",
						Menu: []ScaffoldMenuEntry{
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
