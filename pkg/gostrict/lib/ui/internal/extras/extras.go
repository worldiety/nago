// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package extras is a program, which uses the picker, a background, embedded images, icons and the colors of a
// chart, for the test of the package.
package extras

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	icons "go.wdy.de/nago/pkg/gostrict/lib/ui/icons/flowbite/outline"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

var fruits = list.Freeze(list.Of("Apfel", "Birne", "Kirsche"))

// Render renders the window.
func Render(wnd ui.Window) ui.View {
	picked := ui.IndicesStateOf(wnd, "picked")
	names := list.New[string]()
	for i := range picked.Get().Values() {
		f, _ := fruits.Get(i)
		names.Append(f)
	}

	return ui.VStack(
		ui.Text("gewählt: "+text.Join(names, ", ")),
		ui.PickerFromSlice("Obst", fruits.Len(), picked).
			MultiSelect(true).
			Stringer(func(row int32) string {
				f, _ := fruits.Get(row)
				return f
			}),
		ui.HStack(ui.Text("Verlauf")).
			Background(ui.Background().AppendLinearGradient(ui.A0, ui.M0)).
			Frame(ui.Frame{Height: ui.L48}.FullWidth()),
		ui.Image().Embed(icons.Star).Frame(ui.Frame{}.Size(ui.L24, ui.L24)),
		ui.PrimaryButton(func() { picked.Set(list.Of[int32](0, 2)) }).Title("Speichern").PreIcon(icons.FloppyDisk),
		ui.BarChart(ui.Chart{XAxisTitle: "Monat"}, "Jan", "Feb").
			Series(ui.ChartSeries("Umsatz", ui.ChartSeriesTypeColumn, ui.ChartDataPoint{X: "Jan", Y: 3}, ui.ChartDataPoint{X: "Feb", Y: 5})).
			Colors("#ff0000"),
	).Gap(ui.L8)
}
