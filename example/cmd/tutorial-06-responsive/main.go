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

//go:embed ora_image_black.svg
var imgData application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		oraImgUri := cfg.Resource(imgData)

		cfg.Component(".", func(wnd core.Window) core.View {

			return ui.VStack(
				ui.Text(fmt.Sprintf("size class %s", wnd.Info().SizeClass)),
				ui.ViewThatMatches(wnd,
					ui.SizeClass(ora.SizeClass2XL, largeLayout(oraImgUri)),
					ui.SizeClass(ora.SizeClassSmall, smallLayout()),
					ui.SizeClass(ora.SizeClassMedium, mediumLayout(oraImgUri)),
				),
			).BackgroundColor("#F5F5F5").
				Frame(ora.Frame{}.FullWidth()).Padding(ora.Padding{}.All(ora.L44))
		})
	}).Run()
}

func largeLayout(img ora.URI) core.View {
	return ui.Grid(
		ui.GridCell(heroCard(img)),
		ui.GridCell(heroCard(img)),
		ui.GridCell(heroCard(img)),
	).Rows(1).ColGap(ora.L16)
}

func mediumLayout(img ora.URI) core.View {
	return ui.Grid(
		ui.GridCell(heroCard(img)),
		ui.GridCell(heroCard(img)),
		ui.GridCell(heroCard(img)),
	).Rows(2).Gap(ora.L16)
}

func smallLayout() core.View {
	return ui.Grid(
		ui.GridCell(heroCard("")),
		ui.GridCell(heroCard("")),
		ui.GridCell(heroCard("")),
	).Rows(3).Gap(ora.L16)
}

func heroCard(img ora.URI) core.DecoredView {
	return ui.VStack(
		ui.If(img != "", ui.Image().
			URI(img).
			Border(ora.Border{}.Radius(ora.L16)).
			Frame(ora.Frame{}.Size(ora.Full, "278dp"))),
		ui.VStack(
			ui.VStack(
				ui.Text("Überschrift").Font(ora.Title),
				ui.HLine(),
			),
			ui.Text(`Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.`),
			ui.HStack(
				ui.Text("Standard").
					Border(ora.Border{}.Circle().Width("1dp")).
					Padding(ora.Padding{}.Vertical(ora.L8).Horizontal(ora.L16)),
			).
				Alignment(ora.Trailing).
				Frame(ora.Frame{}.FullWidth()),
		).Alignment(ora.Leading).
			Gap(ora.L16).
			Padding(ora.Padding{}.All(ora.L16)),
	).Alignment(ora.Top).
		BackgroundColor("#FAFAFA").
		Frame(ora.Frame{}.Size("476dp", ora.Auto)).
		Border(ora.Border{}.Radius(ora.L16))
}
