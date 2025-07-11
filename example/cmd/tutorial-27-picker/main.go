// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/picker"
	"go.wdy.de/nago/web/vuejs"
)

type Person struct {
	Vorname, Nachname string
}

func (p Person) String() string {
	return fmt.Sprintf("%s %s", p.Vorname, p.Nachname)
}

var names = []string{"Baba", "Noah", "Ethan", "Olivia", "Isabella", "Jacob", "Ava", "Liam", "Logan", "Sophia", "Emily", "Michael", "Madison", "Matthew", "Jack", "Mia", "Hannah", "Ryan", "Abigail"}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		var persons []Person
		for _, first := range names {
			for _, second := range names {
				persons = append(persons, Person{first, second})
			}
		}

		cfg.RootView(".", func(wnd core.Window) core.View {
			enabled := core.AutoState[bool](wnd)
			personState := core.AutoState[[]Person](wnd).Init(func() []Person {
				return []Person{persons[5]}
			})
			personState.Observe(func(newValue []Person) {
				enabled.Set(len(newValue) > 0)
			})

			err := std.NewLocalizedError("hello", "hello world")

			return VStack(
				picker.Picker[Person]("Personen", persons, personState).
					SupportingText("Wähle jemanden aus").
					Title("Alle Personen").
					MultiSelect(true).
					DialogOptions(
						alert.Height("20dvh"),
					).
					//ErrorText("Falsch").
					Frame(Frame{Width: L320}),
				PrimaryButton(func() {
					fmt.Println(personState)
				}).Title("print selected").Enabled(enabled.Get()),
				alert.BannerError(err),
			).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
