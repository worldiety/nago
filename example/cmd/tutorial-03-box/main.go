package main

import (
	_ "embed"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed profile.jpg
var profileData application.StaticBytes

//go:embed gras.jpg
var grasData application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		profileURI := cfg.Resource(profileData)
		grassURI := cfg.Resource(grasData)

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(
				Card(
					ui.HStack(
						Avatar(profileURI),
						Details("Sir Gopher", "3 minutes ago"),
					),
					PostedImage(grassURI),
				),
			).Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}

func Avatar(data ora.URI) core.View {
	return ui.Box(ui.BoxLayout{
		Center: ui.Image().
			URI(data).
			Frame(ora.Frame{}.Size(ora.L120, ora.L120)).
			Border(ora.Border{}.Circle().Width(ora.L4).Color("#ffffff").Shadow(ora.L4)),
		BottomTrailing: ui.Box(ui.BoxLayout{
			Center: ui.Text("42").
				Font(ora.Font{Weight: ora.BoldFontWeight}).
				Color("#2d6187"),
		}).
			BackgroundColor("#52eb8f").
			Border(ora.Border{}.Circle().Width(ora.L4).Color("#ffffff")).
			Frame(ora.Frame{}.Size(ora.L44, ora.L44)),
	}).
		Frame(ora.Frame{}.Size(ora.L120, ora.L120))
}

func PostedImage(data ora.URI) core.View {
	return ui.Image().
		URI(data).
		Frame(ora.Frame{}.Size(ora.Full, ora.Auto)).
		Border(ora.Border{}.Radius(ora.L4).Elevate(2))
}

func Details(headline, subheadline string) core.View {
	return ui.VStack(
		ui.Text(headline).Font(ora.Title),
		ui.Text(subheadline),
	).Alignment(ora.Leading).
		Padding(ora.Padding{}.Horizontal(ora.L20))
}

func Card(views ...core.View) core.View {
	return ui.VStack(views...).
		Gap(ora.L12).
		Alignment(ora.Leading).
		Border(ora.Border{}.Radius(ora.L4).Elevate(4)).
		Frame(ora.Frame{}.Size(ora.L320, ora.Auto)).
		Padding(ora.Padding{}.All(ora.L8))
}
