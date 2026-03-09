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
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dropdown"
	"go.wdy.de/nago/web/vuejs"
)

type ID string
type Person struct {
	ID                ID
	Vorname, Nachname string
}

func (p Person) Identity() ID {
	return p.ID
}

func (p Person) String() string {
	return fmt.Sprintf("%s %s", p.Vorname, p.Nachname)
}

var names = []string{"Baba", "Noah", "Ethan", "Olivia", "Isabella", "Jacob", "Ava", "Liam", "Logan", "Sophia", "Emily", "Michael", "Madison", "Matthew", "Jack", "Mia", "Hannah", "Ryan", "Abigail"}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_89")
		cfg.Serve(vuejs.Dist())

		var persons []Person
		for i, first := range names {
			for j, second := range names {
				persons = append(persons, Person{
					ID:       ID(fmt.Sprintf("%d-%d-%s-%s", i, j, first, second)),
					Vorname:  first,
					Nachname: second,
				})
			}
		}

		selectOptions := make([]dropdown.Option[ID], 0, len(persons))
		for i, person := range persons {
			selectOptions = append(selectOptions, dropdown.Option[ID]{
				Label:    person.String(),
				Value:    person.ID,
				Disabled: i%10 == 0,
			})
		}

		cfg.RootView(".", func(wnd core.Window) core.View {
			personState := core.AutoState[ID](wnd).Init(func() ID {
				return persons[4].ID
			})
			personState.Observe(func(newValue ID) {
				fmt.Println("Select value changed: " + personState.Get())
			})

			return ui.VStack(
				dropdown.Dropdown("Person auswählen", selectOptions, personState.Get()).InputValue(personState),
				ui.PrimaryButton(func() {
					fmt.Println(personState)
				}).Title("print selected").Enabled(len(personState.Get()) > 0),
				ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo("picker-drop-in", nil)
				}).Title("show picker-drop-in api"),
			).
				Gap(ui.L16).
				Frame(ui.Frame{}.MatchScreen())
		})

		cfg.RootView("picker-drop-in", func(wnd core.Window) core.View {
			personState := core.AutoState[[]Person](wnd).Init(func() []Person {
				return []Person{persons[4]}
			})
			personState.Observe(func(newValue []Person) {
				fmt.Println("Select value changed: ", personState.Get())
			})

			return ui.VStack(
				dropdown.FromSlice("Person auswählen", persons, personState),
				ui.PrimaryButton(func() {
					fmt.Println(personState)
				}).Title("print selected").Enabled(len(personState.Get()) > 0),
			).
				Gap(ui.L16).
				Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
