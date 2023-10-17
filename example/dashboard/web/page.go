package web

import (
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/presentation/ui"
	"net/http"
)

type DashboardModel struct {
	Title string
	Count int
}

type AddEvent int
type SubEvent int

func Home() http.HandlerFunc {
	return ui.Handler(
		Render,
		ui.OnEvent(func(model DashboardModel, evt AddEvent) DashboardModel {
			model.Count++
			return model
		}),
		ui.OnEvent(func(model DashboardModel, evt SubEvent) DashboardModel {
			model.Count--
			return model
		}),
	)
}

func Render(model DashboardModel) ui.Page {
	return ui.Page{
		Title: model.Title,
		Body: ui.Grid{
			Columns: 3,
			Cells: slice.Of(
				ui.GridCell{
					Span: 3,
					Views: slice.Of[ui.View](
						ui.Form{Views: slice.Of[ui.InputType](
							ui.InputFile{
								Name:     "UnsafeName",
								Multiple: false,
							},
						)},
					),
				},
				ui.GridCell{
					Span: 2,
				},
				ui.GridCell{
					Views: slice.Of[ui.View](
						ui.Button{
							Title:   ui.AttributedText{Value: "Plus"},
							OnClick: AddEvent(1),
						},
					),
				},
				ui.GridCell{Views: slice.Of[ui.View](
					ui.Button{
						Title:   ui.AttributedText{Value: "Minus"},
						OnClick: SubEvent(1),
					},
				)},
			),
		},
	}
}
