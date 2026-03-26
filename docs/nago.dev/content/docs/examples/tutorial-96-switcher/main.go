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
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
	. "go.wdy.de/nago/presentation/ui/switcher"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_96")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			pages := []TSwitcherPage{
				SwitcherPage(
					"switcher_page_1",
					"Switcher Seite 1",
					icons.Banknotes,
					VStack(
						Text("Lorem Ipsum").Font(HeadlineMedium),
						Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum."),
						HStack(
							Button(ButtonStylePrimary, func() {}).Title("Dummy Button").FullWidth(),
						).FullWidth().Padding(Padding{Top: L8}),
					).Alignment(BottomLeading),
				).Img("https://picsum.photos/300/600"),
				SwitcherPage(
					"switcher_page_2",
					"Switcher Seite 2",
					icons.Heart,
					VStack(
						Text("Lorem Ipsum").Font(HeadlineMedium),
						Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua."),
						HStack(
							Button(ButtonStylePrimary, func() {}).Title("Dummy Button").FullWidth(),
						).FullWidth().Padding(Padding{Top: L8}),
					).Alignment(BottomLeading),
				).Img("https://picsum.photos/1200/900"),
				SwitcherPage(
					"switcher_page_3",
					"Switcher Seite 3",
					icons.DocumentText,
					VStack(
						Text("Lorem Ipsum").Font(HeadlineMedium),
						Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua."),
						HStack(
							Button(ButtonStylePrimary, func() {}).Title("Dummy Button"),
						).FullWidth().Alignment(Center).Padding(Padding{Top: L8}),
					).Alignment(BottomLeading),
				),
			}

			return VStack(
				VStack(
					Text("Switcher").Font(HeadlineMedium),
					Switcher(pages, core.AutoState[string](wnd)).Frame(Frame{MaxWidth: "1000px"}).FullWidth().ContentNoPadding(),
				).FullWidth(),
				VStack(
					Text("Switcher mit Object-Fit").Font(HeadlineMedium),
					Switcher(pages, core.AutoState[string](wnd)).ImageObjectFit(FitCover).Frame(Frame{MaxWidth: "1000px"}).FullWidth().ContentNoPadding(),
				).FullWidth(),
				VStack(
					Text("Switcher mit variabler Höhe").Font(HeadlineMedium),
					Switcher(pages, core.AutoState[string](wnd)).Frame(Frame{MaxWidth: "1000px"}).FullWidth().ContentNoPadding().DynamicHeight(),
				).FullWidth(),
			).FullWidth().Alignment(Center).Gap(L32).Frame(Frame{}.FullWidth()).Padding(Padding{Top: L32, Right: L32, Bottom: L320, Left: L32})
		})

	}).Run()
}
