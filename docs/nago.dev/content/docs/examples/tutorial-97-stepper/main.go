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

			return ui.VStack(
				ui.Text("Aktueller Step: "+strconv.Itoa(state.Get()+1)),
				stepper.Stepper(
					stepper.Step().Headline("first").SupportingText("the first thing"),
					stepper.Step().Headline("second").SupportingText("the second thing"),
					stepper.Step().Headline("third").SupportingText("the third thing"),
					stepper.Step().Headline("fourth").SupportingText("the fourth thing"),
					stepper.Step().Headline("fifth").SupportingText("the fifth thing"),
				).InputValue(state),
				ui.Stack(
					ui.PrimaryButton(func() {
						state.Set(state.Get() - 1)
					}).Title("< Zurück").Disabled(state.Get() <= 0),
					ui.PrimaryButton(func() {
						state.Set(state.Get() + 1)
					}).Title("Weiter >").Disabled(state.Get() >= 4),
				).Gap(ui.L8),
			).Gap(ui.L32).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}
