// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"slices"
	"strings"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	names := []string{"Anna", "Ben", "Clara", "David", "Emma", "Felix", "Greta", "Hannah", "Ian", "Julia", "Karl", "Lena", "Mia", "Noah", "Oskar", "Paula", "Quentin", "Rosa", "Samuel", "Tina", "Uwe", "Vera", "Willi", "Xenia", "Yara", "Zoe", "Adrian", "Bianca", "Chris", "Diana", "Elena", "Fabian", "Gustav", "Helena", "Isabel", "Jonas", "Klara", "Levi", "Mara", "Nico", "Oliver", "Pia", "Rafael", "Sara", "Tim", "Ulrike", "Valentin", "Wanda", "Xaver", "Yvonne", "Aaron", "Bruno", "Celine", "Dominik", "Eva", "Franz", "Gina", "Henry", "Inga", "Jannis", "Kilian", "Lara", "Matteo", "Nina", "Olivier", "Petra", "Rico", "Selina", "Tobias", "Udo", "Vanessa", "Walter", "Ximena", "Yusuf", "Zara", "Alina", "Bastian", "Carla", "Dennis", "Elisa", "Fiona", "Georg", "Helge", "Ilona", "Jule", "Konrad", "Lukas", "Milan", "Nora", "Otto", "Philipp", "Ronja", "Sven", "Thea", "Urs"}

	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			text := core.AutoState[string](wnd).Init(func() string {
				return ""
			})

			return VStack(
				VStack(
					TextField("Filter", text.Get()).
						InputValue(text).
						FullWidth(),
					ScrollView(
						VStack().Append(
							slices.Collect(func(yield func(core.View) bool) {
								filteredNames := make([]string, 0)
								for _, name := range names {
									if strings.Contains(strings.ToLower(name), strings.ToLower(text.Get())) {
										filteredNames = append(filteredNames, name)
									}
								}
								for _, name := range filteredNames {
									yield(Text(name))
								}
							})...,
						),
					).
						Frame(Frame{MaxHeight: L200}),
				).
					Gap(L16).
					Frame(Frame{MaxWidth: L880}),
			).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
