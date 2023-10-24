package web

import (
	. "go.wdy.de/nago/presentation/ui"
)

type DashboardModel struct {
	Title string
	Count int
}

type AddEvent int
type SubEvent struct {
	UnsafeName string
	Vorname    string
}

func Home() PageHandler {
	return Page(
		"Hello Page",
		Render,
		OnEvent(func(model DashboardModel, evt AddEvent) DashboardModel {
			model.Count++
			return model
		}),
		OnEvent(func(model DashboardModel, evt SubEvent) DashboardModel {
			model.Count--
			return model
		}),
	)
}

func Render(model DashboardModel) View {
	return Text("hallo welt")
}
