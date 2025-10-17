// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	cfglocalization "go.wdy.de/nago/application/localization/cfg"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed landscape1.png
var landscape1 application.StaticBytes

//go:embed landscape2.png
var landscape2 application.StaticBytes

//go:embed landscape3.png
var landscape3 application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_78")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.Must(cfginspector.Enable(cfg))
		option.Must(cfglocalization.Enable(cfg))

		headlineLandscape1 := "Headline 1"
		textLandscape1 := "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."
		landscape1Content := ui.Content{
			ID:       "landscape1",
			Image:    landscape1,
			Icon:     icons.Bookmark,
			Headline: headlineLandscape1,
			Text:     textLandscape1,
		}

		headlineLandscape2 := "Headline 2"
		textLandscape2 := "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren,"
		landscape2Content := ui.Content{
			ID:       "landscape2",
			Image:    landscape2,
			Icon:     icons.CheckCircle,
			Headline: headlineLandscape2,
			Text:     textLandscape2,
		}

		headlineLandscape3 := "Headline 3"
		textLandscape3 := "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.\n\nLorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At"
		landscape3Content := ui.Content{
			ID:       "landscape3",
			Image:    landscape3,
			Icon:     icons.Rocket,
			Headline: headlineLandscape3,
			Text:     textLandscape3,
		}

		contents := []ui.Content{landscape1Content, landscape2Content, landscape3Content}

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {

			selectedIdx := core.AutoState[int](wnd)

			return ui.VStack(
				ui.Switcher(selectedIdx, contents...).
					ID("switcher").
					AccessibilityLabel("test-label"),
			).Padding(ui.Padding{Top: ui.L80}).Frame(ui.Frame{}.FullWidth())

		})
	}).Run()
}
