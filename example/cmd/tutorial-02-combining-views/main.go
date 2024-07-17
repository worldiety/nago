package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed hummel.jpg
var hummelData application.Bytes

//go:embed gras.jpg
var grasData application.Bytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		hummelUri := cfg.Resource(hummelData)
		grasUri := cfg.Resource(grasData)

		cfg.Component(".", func(wnd core.Window) core.View {
			fmt.Println("hello")
			return ui.VStack(
				ui.Image().
					URI(grasUri).
					Frame(ora.Frame{}.Size("", ora.L320)),

				CircleImage(hummelUri).
					AccessibilityLabel("Hummel an Lavendel").
					Padding(ora.Padding{Top: ora.L160.Negate()}),

				ui.VStack(
					ui.Text("Hummel").
						Font(ora.Title),

					ui.HStack(
						ui.Text("WZO Terrasse"),
						ui.Spacer(),
						ui.Text("Oldenburg"),
					).Frame(ora.Frame{}.FullWidth()).
						Font(ora.Font{Size: ora.L12}),

					ui.HDivider(),
					ui.Text("Es gibt auch").Font(ora.Title),
					ui.Text("Andere Viecher"),
				).Alignment(ora.Leading).
					Frame(ora.Frame{Width: ora.L320}),
			).
				Frame(ora.Frame{Height: ora.ViewportHeight, Width: ora.Full})

		})

	}).Run()
}

func CircleImage(data ora.URI) core.DecoredView {
	return ui.Image().
		URI(data).
		Border(ora.Border{}.
			Shadow(ora.L8).
			Color("#ffffff").
			Width(ora.L4).
			Circle()).
		Frame(ora.Frame{}.Size(ora.L320, ora.L320))

}
