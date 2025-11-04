// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// main denotes an executable go package. If you don't know, what that means, go through the Go Tour first.
package main

import (
	"fmt"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfgent "go.wdy.de/nago/application/ent/cfg"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

type PersonID string

type Person struct {
	ID         PersonID   `json:"id,omitempty" visible:"false"`
	Firstname  string     `json:"firstname,omitempty"`
	Lastname   string     `json:"lastname,omitempty"`
	Age        int        `json:"age,omitempty"`
	BestFriend PersonID   `json:"best,omitempty" source:"testdomain.person"`
	Friends    []PersonID `json:"friends,omitempty" source:"testdomain.person"`
	Hobbies    []HobbyID  `json:"hobbies,omitempty" source:"testdomain.hobby"`
}

func (p Person) Identity() PersonID {
	return p.ID
}

func (p Person) WithIdentity(id PersonID) Person {
	p.ID = id
	return p
}

func (p Person) String() string {
	return fmt.Sprintf("%s %s", p.Firstname, p.Lastname)
}

type HobbyID string
type Hobby struct {
	ID          HobbyID `json:"id,omitempty" visible:"false"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty" label:"nago.common.label.description" lines:"3"`
}

func (h Hobby) Identity() HobbyID {
	return h.ID
}

func (h Hobby) WithIdentity(id HobbyID) Hobby {
	h.ID = id
	return h
}

func (h Hobby) String() string {
	return h.Name
}

// the main function of the program, which is like the java public static void main.
func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_21")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Login(true).
			Decorator())

		option.MustZero(cfg.StandardSystems())
		option.Must(cfginspector.Enable(cfg))
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		option.Must(cfgent.Enable(cfg, "testdomain.person", "Person", cfgent.Options[Person, PersonID]{}))
		option.Must(cfgent.Enable(cfg, "testdomain.hobby", "Hobby", cfgent.Options[Hobby, HobbyID]{}))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return VStack(Text("1. create a role\n2.assign all permissions\n3. give this role to yourself\n4.admin center shows Person and Hobby cards")).
				Frame(Frame{}.MatchScreen())
		})
	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}
