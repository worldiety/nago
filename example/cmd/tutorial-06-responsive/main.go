package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed ora_image_black.svg
var imgData application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		oraImgUri := cfg.Resource(imgData)

		cfg.RootView(".", func(wnd core.Window) core.View {

			return VStack(
				Text(fmt.Sprintf("size class %s", wnd.Info().SizeClass)),
				ViewThatMatches(wnd,
					SizeClass(core.SizeClass2XL, func() core.View { return largeLayout(oraImgUri) }),
					SizeClass(core.SizeClassSmall, func() core.View { return smallLayout() }),
					SizeClass(core.SizeClassMedium, func() core.View { return mediumLayout(oraImgUri) }),
				),
			).BackgroundColor("#F5F5F5").
				Frame(Frame{}.FullWidth()).Padding(Padding{}.All(L44))
		})
	}).Run()
}

func largeLayout(img core.URI) core.View {
	return Grid(
		GridCell(heroCard(img)),
		GridCell(heroCard(img)),
		GridCell(heroCard(img)),
	).Rows(1).ColGap(L16)
}

func mediumLayout(img core.URI) core.View {
	return Grid(
		GridCell(heroCard(img)),
		GridCell(heroCard(img)),
		GridCell(heroCard(img)),
	).Rows(2).Gap(L16)
}

func smallLayout() core.View {
	return Grid(
		GridCell(heroCard("")),
		GridCell(heroCard("")),
		GridCell(heroCard("")),
	).Rows(3).Gap(L16)
}

func heroCard(img core.URI) DecoredView {
	return VStack(
		If(img != "", Image().
			URI(img).
			Border(Border{}.Radius(L16)).
			Frame(Frame{}.Size(Full, "278dp"))),
		VStack(
			VStack(
				Text("Überschrift").Font(Title),
				HLine(),
			),
			Text(`Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.`),
			HStack(
				Text("Standard").
					Border(Border{}.Circle().Width("1dp")).
					Padding(Padding{}.Vertical(L8).Horizontal(L16)),
			).
				Alignment(Trailing).
				Frame(Frame{}.FullWidth()),
		).Alignment(Leading).
			Gap(L16).
			Padding(Padding{}.All(L16)),
	).Alignment(Top).
		BackgroundColor("#FAFAFA").
		Frame(Frame{}.Size("476dp", Auto)).
		Border(Border{}.Radius(L16))
}
