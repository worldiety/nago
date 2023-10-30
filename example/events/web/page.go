package web

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
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

type BlaEvent struct {
	Name       string
	CSV        File
	MultiFiles []File
}

func Home(updateUserName func(name string)) PageHandler {
	return Page(
		"hello-page",
		Render,
		OnEvent(func(model DashboardModel, evt BlaEvent) DashboardModel {
			model.Count++
			updateUserName(evt.Name)

			fmt.Println("got", evt.Name)
			fmt.Printf("got %d bytes\n", len(evt.CSV.Data))
			for _, file := range evt.MultiFiles {
				fmt.Println("multi", file.Name)
			}
			return model
		}),
		OnEvent(func(model DashboardModel, evt SubEvent) DashboardModel {
			model.Count--
			return model
		}),
	)
}

func Render(model DashboardModel) View {
	//	return Text("hallo welt")
	return Grid{
		Columns: 2,
		Gap:     Rem(1),
		Cells: slice.Of(
			GridCell{
				Start: 1,
				End:   3,
				Child: Card{Child: Text("hallo welt")},
			},
			GridCell{Child: InputText{
				Name:  "Name",
				Value: "Torben",
				OnMatchError: slice.Of(
					Match{
						Regex:   "[^1-9]",
						Message: "darf keine Zahl enthalten",
					},
				),
			}},
			GridCell{Child: InputFile{Name: "CSV"}},
			GridCell{Child: InputFile{Name: "MultiFiles", Multiple: true, Accept: slice.Of(".csv", ".pdf")}},
			GridCell{Child: Button{OnClick: BlaEvent{}, Title: Text("Klick")}},
		),
	}
}
