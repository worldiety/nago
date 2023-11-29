package web

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
	. "go.wdy.de/nago/presentation/ui"
)

type DashboardModel struct {
	Title    string
	Count    int
	Vorname  string
	Nachname string
	EMail    string
	Redirect
}

type AddEvent int
type SubEvent struct {
	UnsafeName string
}

type BlaEvent struct {
	Vorname, Nachname, EMail string
	CSV                      File
	MultiFiles               []File
}

func Home(updateUserName func(name string)) PageHandler {
	return Page(
		"hello-page",
		Render,
		OnEvent(func(model DashboardModel, evt BlaEvent) DashboardModel {
			model.Count++
			model.Nachname = evt.Nachname
			model.Vorname = evt.Vorname
			model.EMail = evt.EMail

			fmt.Printf("got %d bytes\n", len(evt.CSV.Data))
			for _, file := range evt.MultiFiles {
				fmt.Println("multi", file.Name)
			}

			if model.Vorname == "Test" {
				model.Redirect = Forward("/counter")
			}

			fmt.Println(model)
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
		Cells: slice.Of(
			GridCell{

				Child: Navbar{
					Caption: Text("Super App"),
					MenuItems: Views(
						Button2{
							Title: Text("Konto"),
						},
					),
				},
			},

			GridCell{
				Child: Grid{
					Padding: Rem(2),
					Columns: 2,
					Rows:    4,
					Gap:     Rem(1),
					Cells: slice.Of(

						GridCell{
							Child: InputText{
								Label: "Dein Vorname",
								Name:  "Vorname",
								Value: model.Vorname,
							},
						},
						GridCell{
							Child: InputText{
								Label: "Dein Nachname",
								Name:  "Nachname",
								Value: model.Nachname,
							}},

						GridCell{
							ColSpan: 2,
							RowSpan: 2,
							Child: InputText{
								Label: "Deine Mail",
								Name:  "EMail",
								Value: model.EMail,
							}},

						GridCell{Child: InputFile{Name: "CSV"}},
						GridCell{Child: InputFile{Name: "MultiFiles", Multiple: true, Accept: slice.Of(".csv", ".pdf")}},
						GridCell{Child: Button2{OnClick: BlaEvent{}, Title: Text("Klick")}},
					),
				},
			},
		),
	}

}
