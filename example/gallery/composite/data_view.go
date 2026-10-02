// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"strings"

	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dataview"
)

func init() {
	app.Register("data-view", func(wnd core.Window) core.View {
		type employee struct {
			Name, Team, City string
		}

		employees := []employee{
			{"Ada Lovelace", "Engineering", "London"},
			{"Grace Hopper", "Operations", "New York"},
			{"Alan Turing", "Research", "Manchester"},
			{"Margaret Hamilton", "Engineering", "Boston"},
		}

		type E = dataview.Element[employee]

		return dataview.FromSlice(wnd, employees, []dataview.Field[E]{
			{
				ID:   "name",
				Name: "Name",
				Map:  func(e E) core.View { return Text(e.Value.Name) },
				Comparator: func(a, b E) int {
					return strings.Compare(a.Value.Name, b.Value.Name)
				},
			},
			{
				ID:   "team",
				Name: "Team",
				Map:  func(e E) core.View { return Text(e.Value.Team) },
			},
			{
				ID:   "city",
				Name: "City",
				Map:  func(e E) core.View { return Text(e.Value.City) },
			},
		}).
			Search(true).
			Selection(true).
			CreateAction(func() {
				// open a create dialog
			}).
			Action(func(e E) {
				// open the details of e.Value
			})
	})
}
