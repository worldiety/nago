// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfgrepoview "go.wdy.de/nago/application/inspector/cfg"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"math/rand"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_57")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		option.MustZero(cfg.StandardSystems())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		option.Must(cfgrepoview.Enable(cfg))

		fillRepoStuff(cfg)

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(ui.Text("hello world")).
				Frame(ui.Frame{}.MatchScreen())

		})
	}).
		Run()
}

type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Age       int      `json:"age,omitempty"`
	Hobbies   []string `json:"hobbies,omitempty"`
}

func (p Person) Identity() string { return p.ID }

var firstnames = []string{"Max", "Anna", "Peter", "Julia", "Lukas", "Laura", "Felix", "Sophie", "Tobias", "Marie"}
var lastnames = []string{"Müller", "Schmidt", "Schneider", "Fischer", "Weber", "Meyer", "Wagner", "Becker", "Hoffmann", "Schäfer"}
var hobbies = []string{"Lesen", "Schwimmen", "Radfahren", "Kochen", "Wandern", "Fotografie", "Gärtnern", "Reisen", "Musik", "Zeichnen"}

func generatePersons(n int) []Person {
	persons := make([]Person, n)

	for i := 0; i < n; i++ {
		persons[i] = Person{
			ID:        data.RandIdent[string](),
			Firstname: firstnames[rand.Intn(len(firstnames))],
			Lastname:  lastnames[rand.Intn(len(lastnames))],
			Age:       rand.Intn(50) + 18, // Alter zwischen 18 und 67
			Hobbies:   randomHobbies(),
		}
	}
	return persons
}

func randomHobbies() []string {
	num := rand.Intn(3) + 1 // 1 bis 3 Hobbies
	hobbySet := make(map[string]struct{})
	var selected []string
	for len(selected) < num {
		hobby := hobbies[rand.Intn(len(hobbies))]
		if _, exists := hobbySet[hobby]; !exists {
			hobbySet[hobby] = struct{}{}
			selected = append(selected, hobby)
		}
	}
	return selected
}

func fillRepoStuff(cfg *application.Configurator) {

	addressbook := application.SloppyRepository[Person, string](cfg)
	if option.Must(addressbook.Count()) < 102 {
		for _, person := range generatePersons(102) {
			option.MustZero(addressbook.Save(person))
		}
	}
}
