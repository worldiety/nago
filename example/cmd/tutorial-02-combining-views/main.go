// main denotes an executable go package. If you don't know, what that means, go through the Go Tour first.
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

// the main function of the program, which is like the java public static void main.
func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		themes := ora.Themes{
			Dark: ora.GenerateTheme(
				ora.PrimaryColor(ora.MustParseHSL("#F7A823")),
				ora.SecondaryColor(ora.MustParseHSL("#00FF00")),
				ora.TertiaryColor(ora.MustParseHSL("#0000FF")),
			),
			Light: ora.GenerateTheme(
				ora.PrimaryColor(ora.MustParseHSL("#F7A823")),
				ora.SecondaryColor(ora.MustParseHSL("#00FF00")),
				ora.TertiaryColor(ora.MustParseHSL("#0000FF")),
			),
			Colors: []ora.ColorRule{
				ora.NewColorRule("my-color", "#ff0000", "#00ff00"),
			},
			Lengths: []ora.LengthRule{
				//	ora.NewLengthRule("my-base", "700px", "4rem"),
				ora.NewLengthRule("my-base", "1px", "2rem"),
			},
		}

		cfg.SetThemes(themes)

		hummelUri := cfg.Resource(hummelData)
		grasUri := cfg.Resource(grasData)

		cfg.Component(".", func(wnd core.Window) core.View {
			msg := core.AutoState[string](wnd)
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
						Size("my-base"),

					ui.HStack(
						ui.Text("WZO Terrasse"),
						ui.Spacer(),
						ui.Text("Oldenburg"),
					).Frame(ora.Frame{}.FullWidth()),

					ui.HDivider(),
					ui.Text("Andere Viecher"),
					ui.Text("gibt es auch: "+msg.Get()),
					ui.TextField("Eingabe", msg),
				).Alignment(ora.Leading).
					BackgroundColor("my-color").
					Frame(ora.Frame{Width: "400dp"}),
			).
				Frame(ora.Frame{Height: ora.ViewportHeight, Width: ora.Full})

		})

	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}

type TCircleImage struct {
	core.DecoredView
}

func CircleImage(data ora.URI) TCircleImage {
	return TCircleImage{
		DecoredView: ui.Image().
			URI(data).
			Border(ora.Border{}.
				Shadow(ora.L8).
				Color("#ffffff").
				Width(ora.L4).
				Circle()).
			Frame(ora.Frame{}.Size(ora.L320, ora.L320)),
	}
}
