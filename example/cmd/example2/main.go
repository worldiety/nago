package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"io"
	"log"
)

type PID string

type Person struct {
	ID        PID
	Firstname string
}

func (p Person) Identity() PID {
	return p.ID
}

type PetID int

type Pet struct {
	ID   PetID
	Name string
}

func (p Pet) Identity() PetID {
	return p.ID
}

type IntlIke int

type EntByInt struct {
	ID IntlIke
}

func (e EntByInt) Identity() IntlIke {
	return e.ID
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
				ID:   1,
				Name: "Katze",
			},
			Pet{
				ID:   2,
				Name: "Hund",
			},
			Pet{
				ID:   3,
				Name: "Esel",
			},
			Pet{
				ID:   4,
				Name: "Stadtmusikant",
			},
		)

		testPet, err2 := pets.Find(3)
		if err2 != nil {
			panic(err2)
		}
		fmt.Println(testPet)

		cfg.Serve(vuejs.Dist())

		type OverParams struct {
			WindparkID  string `path:"windpark-id"`
			Windspargel int    `path:"spargel-id"`
		}

		cfg.Index("/jupp")

		cfg.Page(ui.Page[ui.Void]{
			ID:              "jupp",
			Unauthenticated: true,
			Children: slice.Of[ui.Component[ui.Void]](
				ui.ListView[EntByInt, IntlIke, ui.Void]{
					ID: "bla",
					List: func(p ui.Void) (slice.Slice[ui.ListItem[IntlIke]], error) {
						return slice.Of(ui.ListItem[IntlIke]{
							ID:    1,
							Title: "2",
						}), nil

					},
				},
			),
		})

		cfg.Page(ui.Page[OverParams]{
			ID:              "overview",
			Unauthenticated: true,
			Title:           "Übersicht",
			Description:     "Diese Übersicht zeigt alle Stadtmusikanten an. Der Nutzer kann zudem Löschen und in die Detailansicht.",
			Navigation: slice.Of(
				ui.PageNavTarget{
					Target: "overview/42/60", // TODO fix with typesafe params? problem: package cycles in Go
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
					ID: "edit-person22",
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
								Label:    "iCloud nervt",
								Multiple: true,
								Accept:   ".pdf",
							},
							Chooser: ui.SelectField{
								Label:       "Wähle einen",
								SelectedIDs: []string{"2"},
								Hint:        "genau etwas aus Liste",
								Multiple:    true,
								Disabled:    false,
								List: slice.Of(
									ui.SelectItem{
										ID:      "1",
										Caption: "hallo",
									},
									ui.SelectItem{
										ID:      "2",
										Caption: "welt",
									},
								),
							},
						}
					},
					Load: func(form ExampleForm, params OverParams) ExampleForm {
						form.Vorname.Value = "Torben"
						form.Nachname.Value = "Schinke"
						return form
					},

					Delete: ui.FormAction[ExampleForm, OverParams]{
						Title: "löschen",
						Receive: func(form ExampleForm, params OverParams) (ExampleForm, ui.Action) {
							log.Println("jetzt löschen")
							return form, ui.Redirect{
								Target: "/overview/666/666",
							}
						},
					},
					Submit: ui.FormAction[ExampleForm, OverParams]{
						Title: "und ab dafür",
						Receive: func(form ExampleForm, params OverParams) (ExampleForm, ui.Action) {
							fmt.Printf("%+v", form)
							form.Vorname.Error = form.Vorname.Value + " ist falsch"
							for _, file := range form.Avatar.Files {
								buf, err := io.ReadAll(file.Data)
								if err != nil {
									panic(err)
								}
								if len(buf) != int(file.Size) {
									panic("ooops???")
								}
								fmt.Println("upload stimmt: "+file.Name, len(buf))
							}
							fmt.Println("!!!", params.Windspargel)
							return form, nil
							return form, ui.Redirect{
								Target: "/overview/42/42",
							}
						},
					},
				},

				ui.Table[Person, PID, OverParams]{
					ID:          "table-view",
					Description: "Super Tabelle",
					List: func(p OverParams) (slice.Slice[ui.TableRow[PID]], error) {
						return kv.FilterAndMap(persons, nil, func(e Person) ui.TableRow[PID] {
							return ui.TableRow[PID]{
								ID:     e.ID,
								Action: ui.Redirect{Target: ui.Target(e.ID)},
								Cells: slice.Of(
									ui.TableCell{
										Key:   "Super name",
										Value: e.Firstname,
									},
								),
							}
						})
					},
				},

				ui.CardView[OverParams]{
					ID:          "dashboard-cards",
					Description: "Super dashboard somit",
					List: func(p OverParams) (slice.Slice[ui.Card], error) {
						return slice.Of(
							ui.Card{
								Title:       "Helden",
								Subtitle:    "Super low code",
								Content:     ui.CardText("Toller content auf Kachel"),
								PrependIcon: ui.FontIcon{Name: "mdi-check"},
								AppendIcon:  ui.FontIcon{Name: "mdi-account"},
								Actions: slice.Of(
									ui.Button{
										Caption: "Einstieg",
										Action:  ui.Redirect{Target: "/overview/1/2"},
									},
								),
							},

							ui.Card{
								Title:       "Helden2",
								Subtitle:    "Super low code2",
								Content:     ui.CardText("Toller content auf Kachel2"),
								PrependIcon: ui.FontIcon{Name: "mdi-alarm"},
								AppendIcon:  ui.FontIcon{Name: "mdi-airport"},
								Actions: slice.Of(
									ui.Button{
										Caption: "Einstieg",
										Action:  ui.Redirect{Target: "/jupp"},
									},
								),
							},
						), nil
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
	Chooser  ui.SelectField
}
