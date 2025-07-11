// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"slices"
)

var months = []string{"Januar", "Februar", "March", "April", "May", "June", "Juli", "August", "September", "October", "November", "Dezember"}
var names = []string{"Beor der Alte", "Betsy Butterblume", "Bilbo Beutlin", "Adalbert Bolger"}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(gantt()).Frame(Frame{}.FullWidth()).Padding(Padding{}.All(L44))
		})
	}).Run()
}

func gantt() core.View {
	return Grid(
		slices.Collect(func(yield func(cell TGridCell) bool) {
			for _, view := range ganttHeader() {
				yield(GridCell(view))
			}

			for i, name := range names {
				for _, view := range ganttRow(i, name) {
					yield(view)
				}
			}

			yield(vacation())
		})...,
	). // be careful with cols and rows and be better as explicit as possible
		Widths(L160).
		Columns(13).
		RowGap(L8).
		Rows(len(names) + 1).
		Padding(Padding{Bottom: L8}).
		Border(Border{}.TopRadius(L16).Elevate(8)).
		Frame(Frame{}.Size(Full, Auto))
}

func ganttHeader() []core.View {
	return slices.Collect(func(yield func(core.View) bool) {
		yield(headCell("Mitarbeiter"))
		for _, month := range months {
			yield(headCell(month))
		}
	})
}

func ganttRow(idx int, name string) []TGridCell {
	return slices.Collect(func(yield func(cell TGridCell) bool) {
		yield(GridCell(Box(BoxLayout{Center: Text(name)}).BackgroundColor("#ff0000")).
			ColStart(1).
			ColEnd(2),
		)
		yield(GridCell(Box(BoxLayout{
			Leading: Text("verplant")}).
			BackgroundColor("#2ecaac").
			Padding(Padding{Left: L8}).
			Border(Border{}.Circle()),
		). // be careful with cols and rows and be better as explicit as possible
			ColStart(idx*2 + 2).
			ColEnd(idx*2 + 5).
			RowStart(idx + 2).
			RowEnd(idx + 3))

	})
}

func vacation() TGridCell {
	return GridCell(
		Box(BoxLayout{Center: Text("Urlaub").Font(Title).Color("#ffffff")}).
			BackgroundColor("#ff6252aa").Border(Border{}.Radius(L8).Shadow(L16)),
	).
		RowStart(2).
		RowEnd(6).
		ColStart(5).
		ColEnd(9).
		Padding(Padding{}.All(L8))
}

func headCell(text string) core.View {
	return VStack(
		Text(text).Color("#ffffff")).
		BackgroundColor("#0a3444").Frame(Frame{Height: L44})
}
