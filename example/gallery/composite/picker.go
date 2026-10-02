// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
)

type person struct {
	Name string
	Team string
}

func (p person) String() string {
	return p.Name
}

func init() {
	app.Register("picker", func(wnd core.Window) core.View {
		people := []person{
			{"Ada Lovelace", "Engineering"},
			{"Grace Hopper", "Operations"},
			{"Alan Turing", "Research"},
			{"Margaret Hamilton", "Engineering"},
		}

		selected := core.AutoState[[]person](wnd).Init(func() []person {
			return []person{people[0], people[2]}
		})

		// open the dialog right away, as if the user had clicked the field
		presented := core.AutoState[bool](wnd).Init(func() bool { return true })

		return picker.Picker[person]("Reviewers", people, selected).
			WithDialogPresented(presented).
			MultiSelect(true).
			Title("Choose reviewers").
			SupportingText("At least one reviewer is required").
			Frame(Frame{Width: L320})
	})
}
