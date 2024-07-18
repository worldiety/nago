package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

var months = []string{"Januar", "Februar", "March", "April", "May", "June", "Juli", "August", "September", "October", "November", "Dezember"}
var names = []string{"Beor der Alte", "Betsy Butterblume", "Bilbo Beutlin", "Adalbert Bolger"}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(gantt()).Frame(ora.Frame{}.FullWidth()).Padding(ora.Padding{}.All(ora.L44))
		})
	}).Run()
}

func gantt() core.View {
	return ui.Grid(
		slices.Collect(func(yield func(cell ui.TGridCell) bool) {
			for _, view := range ganttHeader() {
				yield(ui.GridCell(view))
			}

			for i, name := range names {
				for _, view := range ganttRow(i, name) {
					yield(view)
				}
			}

			yield(vacation())
		})...,
	). // be careful with cols and rows and be better as explicit as possible
		Widths(ora.L160).
		Columns(13).
		RowGap(ora.L8).
		Rows(len(names) + 1).
		Padding(ora.Padding{Bottom: ora.L8}).
		Border(ora.Border{}.TopRadius(ora.L16).Elevate(8)).
		Frame(ora.Frame{}.Size(ora.Full, ora.Auto))
}

func ganttHeader() []core.View {
	return slices.Collect(func(yield func(core.View) bool) {
		yield(headCell("Mitarbeiter"))
		for _, month := range months {
			yield(headCell(month))
		}
	})
}

func ganttRow(idx int, name string) []ui.TGridCell {
	return slices.Collect(func(yield func(cell ui.TGridCell) bool) {
		yield(ui.GridCell(ui.Box(ui.BoxLayout{Center: ui.Text(name)})).
			ColStart(1).
			ColEnd(2),
		)
		yield(ui.GridCell(ui.Box(ui.BoxLayout{
			Leading: ui.Text("verplant")}).
			BackgroundColor("#2ecaac").
			Padding(ora.Padding{Left: ora.L8}).
			Border(ora.Border{}.Circle()),
		). // be careful with cols and rows and be better as explicit as possible
			ColStart(idx*2 + 2).
			ColEnd(idx*2 + 5).
			RowStart(idx + 2).
			RowEnd(idx + 3))

	})
}

func vacation() ui.TGridCell {
	return ui.GridCell(
		ui.Box(ui.BoxLayout{Center: ui.Text("Urlaub").Font(ora.Title).Color("#ffffff")}).
			BackgroundColor("#ff6252aa").Border(ora.Border{}.Radius(ora.L8).Shadow(ora.L16)),
	).
		RowStart(2).
		RowEnd(6).
		ColStart(5).
		ColEnd(9).
		Padding(ora.Padding{}.All(ora.L8))
}

func headCell(text string) core.View {
	return ui.VStack(
		ui.Text(text).Color("#ffffff")).
		BackgroundColor("#0a3444").Frame(ora.Frame{Height: ora.L44})
}
