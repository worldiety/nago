package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

type PID string

type Person struct {
	ID        PID
	Firstname string
}

func (p Person) Identity() PID {
	return p.ID
}

type PetID string

type Pet struct {
	ID   PetID
	Name string
}

func (p Pet) Identity() PetID {
	return p.ID
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example 2")

		persons := kv.NewCollection[Person, PID](cfg.Store("test2-db"), "persons")
		err := persons.Save(
			Person{
				ID:        "1",
				Firstname: "Frodo",
			},
			Person{
				ID:        "2",
				Firstname: "Sam",
			},
			Person{
				ID:        "3",
				Firstname: "Pippin",
			},
		)
		if err != nil {
			panic(err)
		}

		pets := kv.NewCollection[Pet, PetID](cfg.Store("test2-db"), "pets")
		pets.Save(
			Pet{
				ID:   "1-cat",
				Name: "Katze",
			},
			Pet{
				ID:   "2-dog",
				Name: "Hund",
			},
			Pet{
				ID:   "3-Esel",
				Name: "Esel",
			},
			Pet{
				ID:   "4-blub",
				Name: "Stadtmusikant",
			},
		)

		cfg.Serve(vuejs.Dist())

		type OverParams struct {
			WindparkID  string `path:"windpark-id"`
			Windspargel int    `path:"spargel-id"`
		}

		cfg.Page(ui.Page[OverParams]{
			ID:              "overview",
			Unauthenticated: true,
			Title:           "Übersicht",
			Description:     "Diese Übersicht zeigt alle Stadtmusikanten an. Der Nutzer kann zudem Löschen und in die Detailansicht.",
			Navigation: slice.Of(
				ui.PageNavTarget{
					Target: "overview/42/60", // TODO fix with typesafe params?
					Icon:   ui.FontIcon{Name: "mdi-home"},
					Title:  "Übersicht",
				},
			),
			Children: slice.Of[ui.Component[OverParams]](
				ui.ListView[Person, PID, OverParams]{
					ID:          "personen",
					Description: "Kleine Listenansicht",
					List: func(p OverParams) (slice.Slice[ui.ListItem[PID]], error) {
						return kv.FilterAndMap(persons, nil,
							func(e Person) ui.ListItem[PID] {
								return ui.ListItem[PID]{
									ID:    e.ID,
									Title: e.Firstname,
								}
							},
						)
					},
					Delete: func(p OverParams, ids slice.Slice[PID]) error {
						return persons.Delete(slice.UnsafeUnwrap(ids)...)
					},
				},

				ui.Form[ExampleForm, OverParams]{
					ID: "edit-person",
					Init: func(params OverParams) ExampleForm {
						return ExampleForm{
							Vorname: ui.TextField{
								Label: "v-o-r",
								Hint:  "Dein Rufname",
							},
							Nachname: ui.TextField{
								Label: "Dein Familienname",
								Hint:  "machs besser",
							},
							Avatar: ui.FileUploadField{
								Label: "iCloud nervt",
							},
						}
					},
					Load: func(form ExampleForm, params OverParams) ExampleForm {
						form.Vorname.Value = "Torben"
						form.Nachname.Value = "Schinke"
						return form
					},
					Submit: ui.FormAction[ExampleForm, OverParams]{
						Title: "und ab dafür",
						Receive: func(form ExampleForm, params OverParams) (ExampleForm, ui.Action) {
							fmt.Printf("%+v", form)
							form.Vorname.Error = form.Vorname.Value + " ist falsch"
							return form, nil
						},
					},
				},
			),
		})

	}).Run()
}

type ExampleForm struct {
	Vorname  ui.TextField `class:"col-start-2 col-span-4"`
	Nachname ui.TextField
	Avatar   ui.FileUploadField
}
