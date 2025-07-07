// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
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

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(
				Card(
					HStack(
						Avatar(profileURI),
						Details("Sir Gopher", "3 minutes ago"),
					),
					PostedImage(grassURI),
				),
			).Frame(Frame{}.MatchScreen())
		})
	}).Run()
}

func Avatar(data core.URI) core.View {
	return Box(BoxLayout{
		Center: Image().
			URI(data).
			Frame(Frame{}.Size(L120, L120)).
			Border(Border{}.Circle().Width(L4).Color("#ffffff").Shadow(L4)),
		BottomTrailing: Box(BoxLayout{
			Center: Text("42").
				Font(Font{Weight: HeadlineAndTitleFontWeight}).
				Color("#2d6187"),
		}).
			BackgroundColor("#52eb8f").
			Border(Border{}.Circle().Width(L4).Color("#ffffff")).
			Frame(Frame{}.Size(L44, L44)),
	}).
		Frame(Frame{}.Size(L120, L120))
}

func PostedImage(data core.URI) core.View {
	return Image().
		URI(data).
		Frame(Frame{}.Size(Full, Auto)).
		Border(Border{}.Radius(L4).Elevate(2))
}

func Details(headline, subheadline string) core.View {
	return VStack(
		Text(headline).Font(Title),
		Text(subheadline),
	).Alignment(Leading).
		Padding(Padding{}.Horizontal(L20))
}

func Card(views ...core.View) core.View {
	return VStack(views...).
		Gap(L12).
		Alignment(Leading).
		Border(Border{}.Radius(L4).Elevate(4)).
		Frame(Frame{}.Size(L320, Auto)).
		Padding(Padding{}.All(L8))
}
