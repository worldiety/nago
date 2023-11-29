package web

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/presentation/ui"
)

type CounterModel struct {
	Counter int
}

type IncrementEvent struct{}
type DecrementEvent struct{}

func Counter() ui.PageHandler {
	return ui.Page(
		"counter",
		renderCounter,
		ui.OnEvent(func(model CounterModel, event IncrementEvent) CounterModel {
			model.Counter++
			return model
		}),
		ui.OnEvent(func(model CounterModel, event DecrementEvent) CounterModel {
			model.Counter--
			return model
		}),
	)
}

func renderCounter(model CounterModel) ui.View {
	return ui.Grid{
		Columns: 1,
		Cells: slice.Of(
			ui.GridCell{Child: ui.Text(fmt.Sprintf("%d", model.Counter))},
			ui.GridCell{Child: ui.Button2{
				Title:   ui.Text("+"),
				OnClick: IncrementEvent{},
			}},
			ui.GridCell{Child: ui.Button2{
				Title:   ui.Text("-"),
				OnClick: DecrementEvent{},
			}},
		),
	}
}
