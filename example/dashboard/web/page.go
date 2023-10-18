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
type SubEvent struct {
	UnsafeName string
	Vorname    string
}

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
				Children: Views(
					InputFile{
						Name:     "UnsafeName",
						Multiple: false,
					},
					InputText{
						Name: "Vorname",
						OnMatchError: slice.Of(
							Match{
								Regex:   `abc.*`,
								Message: "Darf nicht mit abc beginnen",
							},

							Match{
								Regex:   `^[`,
								Message: "Darf nicht [ enthalten",
							},
						),
						OnMatchSupporting: slice.Of(
							Match{
								Regex:   "Hello world",
								Message: "Das ist super.",
							},
						),
					},
				),
			},
			GridCell{
				Span: 2,
			},
			GridCell{
				Child: Button{
					Title:   Text("Plus"),
					OnClick: AddEvent(1),
				},
			},
			GridCell{
				Child: Button{
					Title:   AttributedText{Value: "Minus"},
					OnClick: SubEvent{},
				},
			},
		),
	}
}
