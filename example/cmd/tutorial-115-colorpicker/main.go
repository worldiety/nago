// Copyright (c) 2026 worldiety GmbH
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
			color := core.AutoState[Color](wnd).Init(func() Color { return "#eb3f33" })
			opaque := core.AutoState[Color](wnd).Init(func() Color { return "#1a73e8" })
			fieldColor := core.AutoState[Color](wnd).Init(func() Color { return "#6750a4" })

			return VStack(
				Text("Picker").Font(Title),
				colorpicker.Picker("Farbe", color.Get()).
					InputValue(color).
					SupportingText("HSV mit Deckkraft, der Standard").
					Frame(Frame{MaxWidth: L480}.FullWidth()),
				Text("Gewählt: "+string(color.Get())),
				HStack().BackgroundColor(color.Get()).Frame(Frame{}.Size(L80, L40)).Border(Border{}.Radius(L8)),

				Text("Picker in HCT ohne Deckkraft").Font(Title),
				colorpicker.Picker("Farbe", opaque.Get()).
					InputValue(opaque).
					Mode(colorpicker.ModeHCT).
					DisableAlpha(true).
					Frame(Frame{MaxWidth: L480}.FullWidth()),
				Text("Gewählt: "+string(opaque.Get())),

				Text("PickerField").Font(Title),
				colorpicker.PickerField("Akzentfarbe", fieldColor.Get()).
					InputValue(fieldColor).
					SupportingText("Ein Klick öffnet den Picker").
					Frame(Frame{MaxWidth: L320}.FullWidth()),
				Text("Gewählt: "+string(fieldColor.Get())),
				colorpicker.PickerField("Deaktiviert", fieldColor.Get()).
					InputValue(fieldColor).
					Disabled(true).
					Frame(Frame{MaxWidth: L320}.FullWidth()),
			).Gap(L16).
				Alignment(Leading).
				Padding(Padding{}.All(L24)).
				Frame(Frame{}.FullWidth())
		})
	}).Run()
}
