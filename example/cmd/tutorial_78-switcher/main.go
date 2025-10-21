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

//go:embed hummel.jpg
var hummel application.StaticBytes

//go:embed gras.jpg
var gras application.StaticBytes

//go:embed screenshot-01.png
var screenShot application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_78")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.Must(cfginspector.Enable(cfg))
		option.Must(cfglocalization.Enable(cfg))

		headline1 := "Headline 1"
		text1 := "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."
		content1 := ui.Content{
			ID:       "gras",
			Image:    gras,
			Icon:     icons.Globe,
			Headline: headline1,
			Text:     text1,
		}

		headline2 := "Headline 2"
		text2 := "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren,"
		content2 := ui.Content{
			ID:       "hummel",
			Image:    hummel,
			Icon:     icons.CameraPhoto,
			Headline: headline2,
			Text:     text2,
		}

		headline3 := "Headline 3"
		text3 := "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.\n\nLorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At"
		content3 := ui.Content{
			ID:       "combined",
			Image:    screenShot,
			Icon:     icons.FileImage,
			Headline: headline3,
			Text:     text3,
		}

		contents := []ui.Content{content1, content2, content3}

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {

			selectedIdx := core.AutoState[int](wnd)

			return ui.VStack(
				ui.Switcher(selectedIdx, contents...).
					ID("switcher").
					AccessibilityLabel("test-label").
					Padding(ui.Padding{}.All(ui.L16)).
					Frame(ui.Frame{}.FullWidth()),
			).Padding(ui.Padding{Top: ui.L80}).Frame(ui.Frame{}.FullWidth())

		})
	}).Run()
}
