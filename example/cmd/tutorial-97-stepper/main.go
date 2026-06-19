// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"
	"strconv"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/stepper"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_97")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			state := core.StateOf[int](wnd, "stepperState")

			nextText := "Weiter >"
			if state.Get() >= 4 {
				nextText = "Fertig >"
			}

			return ui.VStack(
				ui.PrimaryButton(func() {
					if wnd.Info().ColorScheme == core.Light {
						wnd.SetColorScheme(core.Dark)
					} else {
						wnd.SetColorScheme(core.Light)
					}
				}).Title("Toggle theme"),
				ui.Text("Aktueller Step: "+strconv.Itoa(state.Get()+1)),
				stepper.Stepper(
					stepper.Step().Headline("Schritt 1").SupportingText("Einkaufsliste"),
					stepper.Step().Headline("Schritt 2").SupportingText("Einkaufen"),
					stepper.Step().Headline("Schritt 3").SupportingText("Zutaten schnibbeln"),
					stepper.Step().Headline("Schritt 4").SupportingText("Kochen"),
					stepper.Step().Headline("Schritt 5").SupportingText("Essen und freuen"),
				).InputValue(state),
				ui.Stack(
					ui.PrimaryButton(func() {
						state.Set(state.Get() - 1)
					}).Title("< Zurück").Disabled(state.Get() <= 0),
					ui.PrimaryButton(func() {
						state.Set(state.Get() + 1)
					}).Title(nextText).Disabled(state.Get() > 4),
				).Gap(ui.L8),
			).Gap(ui.L32).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}
