// main denotes an executable go package. If you don't know, what that means, go through the Go Tour first.
package main

import (
	_ "embed"
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

		hummelUri := cfg.Resource(hummelData)
		grasUri := cfg.Resource(grasData)

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(func(vstack *ui.ViewVStack) {
				vstack.Frame(ora.Frame{Height: ora.ViewportHeight, Width: ora.Full})

				vstack.Of(
					ui.Image(func(img *ui.ImageView) {
						img.URI(grasUri)
						img.SetFrame(ora.Frame{}.Size("", "300dp"))
					}),
					NewCircleImage(hummelUri, func(img *ui.ImageView) {
						pad := img.Padding()
						pad.Top = "-150dp"
						img.SetPadding(pad)
					}),
					ui.TextFrom("hello world"),
				)
			})

		})
	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}

type CircleImage struct {
	*ui.ImageView
}

func NewCircleImage(data ora.URI, with func(img *ui.ImageView)) *CircleImage {
	return &CircleImage{
		ImageView: ui.Image(func(img *ui.ImageView) {
			img.URI(data)
			img.AccessibilityLabel("Hummel an Lavendel")
			img.Border(
				ora.Border{}.
					Shadow("7dp").
					Color("#ffffff").
					Width("2dp").
					Radius("20dp").
					Circle().
					Width("3dp"),
			)
			img.SetFrame(ora.Frame{}.Size("300dp", "300dp"))

			if with != nil {
				with(img)
			}
		}),
	}
}
