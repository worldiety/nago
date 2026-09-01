// Copyright (c) 2026 worldiety GmbH
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
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dropdown"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_112")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			genderOptions := []dropdown.Option[string]{
				{Value: "male", Label: "Herr"},
				{Value: "female", Label: "Frau"},
				{Value: "diverse", Label: "Divers"},
			}

			stateGender := core.StateOf[string](wnd, "stateGender")
			stateFirstName := core.StateOf[string](wnd, "stateFirstName")
			stateLastName := core.StateOf[string](wnd, "stateLastName")
			stateEmail := core.StateOf[string](wnd, "stateEmail")
			stateComment := core.StateOf[string](wnd, "stateComment")

			gridCols := 2
			if wnd.Info().SizeClass < core.SizeClassMedium {
				gridCols = 1
			}

			return ui.Stack(
				ui.VStack(
					ui.Stack(
						ui.ThemeSwitcher(
							ui.PrimaryButton(nil).Title("Toggle theme"),
						),
					).FullWidth().Padding(ui.Padding{Bottom: ui.L32}),
					ui.Form(
						form.Fieldset(
							ui.Grid(
								ui.GridCell(
									dropdown.Dropdown("Anrede", genderOptions, stateGender.Get()).InputValue(stateGender),
								),
								ui.GridCell(
									ui.TextField("Vorname", stateFirstName.Get()).InputValue(stateFirstName),
								).ColStart(1),
								ui.GridCell(
									ui.TextField("Nachname", stateLastName.Get()).InputValue(stateLastName),
								),
								ui.GridCell(
									ui.TextField("E-Mail", stateEmail.Get()).InputValue(stateEmail).Optional(true),
								).ColSpan(2),
							).Columns(gridCols).Gap(ui.L16),
						).Title("Persönliche Daten"),
						ui.Space(ui.L16),
						form.Fieldset(
							ui.Grid(
								ui.GridCell(
									ui.TextField("Kommentar", stateComment.Get()).InputValue(stateComment).Lines(3).Optional(true),
								).ColSpan(2),
							).Columns(gridCols).Gap(ui.L16),
						).Title("Weiteres"),
					),
				).Alignment(ui.Leading).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{MaxWidth: ui.L1600}),
			).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
