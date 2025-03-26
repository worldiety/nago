// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/colorpicker"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			color := core.AutoState[Color](wnd)

			return VStack(
				colorpicker.PalettePicker("Deine Lieblingsfarbe", colorpicker.DefaultPalette).Value(color.Get()).State(color).
					Title("Bitte Farbe wählen"),
			).Gap(L16).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
