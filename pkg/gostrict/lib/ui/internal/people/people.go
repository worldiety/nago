// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package people is a program, which shows its rows in a data view, for the test of the package.
package people

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Person is a row.
type Person struct {
	Name string
	Age  int32
}

var people = list.Freeze(list.Of(
	Person{Name: "Carla", Age: 41},
	Person{Name: "Anna", Age: 29},
	Person{Name: "Bob", Age: 35},
))

// Render renders the window.
func Render(wnd ui.Window) ui.View {
	chosen := ui.StringStateOf(wnd, "chosen")
	return ui.VStack(
		ui.Text("gewählt: "+chosen.Get()),
		ui.DataViewFromSlice(wnd, people.Len(),
			ui.DataViewField("Name", func(row int32) ui.View {
				p, _ := people.Get(row)
				return ui.Text(p.Name)
			}).ID("name").Comparator(func(a, b int32) int32 {
				pa, _ := people.Get(a)
				pb, _ := people.Get(b)
				return text.Compare(pa.Name, pb.Name)
			}),
			ui.DataViewField("Alter", func(row int32) ui.View {
				p, _ := people.Get(row)
				return ui.Text(text.FromInt(p.Age))
			}).ID("age").MinSize(ui.SizeClassMedium),
		).
			Style(ui.DataViewTable).
			ColumnWidth("age", ui.L80).
			Action(func(row int32) {
				p, _ := people.Get(row)
				chosen.Set(p.Name)
			}).
			SelectOptions(ui.DataViewSelectOptionDelete(wnd, func(rows list.List[int32]) error {
				chosen.Set("gelöscht " + text.FromInt(rows.Len()))
				return nil
			})),
	)
}
