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
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

type Person struct {
	ID                string
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
		for i, first := range names {
			for j, second := range names {
				persons = append(persons, Person{
					ID:       fmt.Sprintf("%d-%d-%s-%s", i, j, first, second),
					Vorname:  first,
					Nachname: second,
				})
			}
		}

		selectOptions := make([]SelectOption, 0, len(persons))
		for i, person := range persons {
			selectOptions = append(selectOptions, SelectOption{
				Label:    person.String(),
				Value:    person.ID,
				Disabled: i%10 == 0,
			})
		}

		cfg.RootView(".", func(wnd core.Window) core.View {
			enabled := core.AutoState[bool](wnd)
			personState := core.AutoState[string](wnd)
			personState.Observe(func(newValue string) {
				fmt.Println("VALUE CHANGED: " + personState.Get())
			})

			return VStack(
				Select(selectOptions).Label("Person auswählen").InputValue(personState),
				PrimaryButton(func() {
					fmt.Println(personState)
				}).Title("print selected").Enabled(enabled.Get()),
			).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
