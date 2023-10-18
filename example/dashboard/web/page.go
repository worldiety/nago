package web

import (
	"go.wdy.de/nago/container/slice"
	. "go.wdy.de/nago/presentation/ui"
	"net/http"
)

type DashboardModel struct {
	Title string
	Count int
}

type AddEvent int
type SubEvent int

func Home() http.HandlerFunc {
	return Handler(
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
	return Grid{
		Columns: 3,
		Cells: slice.Of(
			GridCell{
				Span: 3,
				Views: Views(
					InputFile{
						Name:     "UnsafeName",
						Multiple: false,
					},
				),
			},
			GridCell{
				Span: 2,
			},
			GridCell{
				Views: Views(
					Button{
						Title:   Text("Plus"),
						OnClick: AddEvent(1),
					},
				),
			},
			GridCell{Views: Views(
				Button{
					Title:   AttributedText{Value: "Minus"},
					OnClick: SubEvent(1),
				},
			)},
		),
	}
}
