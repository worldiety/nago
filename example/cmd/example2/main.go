package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"io"
	"log"
	"time"
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

type MyForm struct {
	Name    ui.TextField
	Check   ui.SwitchField
	DueDate ui.DateField
	Comment ui.TextAreaField
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example 2")

		cfg.KeycloakAuthentication()
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
			User        auth.User
		}

		cfg.Index("/jupp")

		cfg.Page(ui.Page[ui.Void]{
			ID:              "jupp",
			Unauthenticated: true,
			Children: slice.Of[ui.Component[ui.Void]](
				ui.SVG[ui.Void]{
					ID:       "example-svg",
					MaxWidth: 344,
					Render: func(p ui.Void) (string, error) {
						return `<svg width="100%" height="100%" viewBox="0 0 7016 9934" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xml:space="preserve" xmlns:serif="http://www.serif.com/" style="fill-rule:evenodd;clip-rule:evenodd;stroke-linejoin:round;stroke-miterlimit:2;">

<style>
    .bar:hover{
      fill:red;
	background-color:yellow;
    }

	.bar{
      fill: rgb(235,235,235);
	}

	.foo:hover{
		fill:green;
	}

	.foo{
		fill: rgb(235,235,235);
	}
   
  </style>

    <g>
		<a href="/overview/42/42">
			<g class="bar">
				<g transform="matrix(1,0,0,1,-11.6512,467.21)">
					<path  d="M5890.31,3829.26C5890.31,3560.02 5671.72,3341.43 5402.47,3341.43L2109.78,3341.43C1840.53,3341.43 1621.94,3560.02 1621.94,3829.26L1621.94,4804.93C1621.94,5074.18 1840.53,5292.77 2109.78,5292.77L5402.47,5292.77C5671.72,5292.77 5890.31,5074.18 5890.31,4804.93L5890.31,3829.26Z" />
				</g>
				<g transform="matrix(1,0,0,1,-11.6512,3590.06)">
					<path d="M5890.31,3829.26C5890.31,3560.02 5671.72,3341.43 5402.47,3341.43L2109.78,3341.43C1840.53,3341.43 1621.94,3560.02 1621.94,3829.26L1621.94,4804.93C1621.94,5074.18 1840.53,5292.77 2109.78,5292.77L5402.47,5292.77C5671.72,5292.77 5890.31,5074.18 5890.31,4804.93L5890.31,3829.26Z" />
				</g>
			</g>
		</a>
		<a href="/overview/43/43">
			<g class="foo">
				<g transform="matrix(1.26344,0,0,0.280575,-1563.41,2106.93)">
					<path d="M5890.31,3829.26C5890.31,3560.02 5841.77,3341.43 5781.98,3341.43L1730.27,3341.43C1670.48,3341.43 1621.94,3560.02 1621.94,3829.26L1621.94,4804.93C1621.94,5074.18 1670.48,5292.77 1730.27,5292.77L5781.98,5292.77C5841.77,5292.77 5890.31,5074.18 5890.31,4804.93L5890.31,3829.26Z" />
				</g>
				<g transform="matrix(1.26344,0,0,0.280575,-1563.41,5229.78)">
					<path d="M5890.31,3829.26C5890.31,3560.02 5841.77,3341.43 5781.98,3341.43L1730.27,3341.43C1670.48,3341.43 1621.94,3560.02 1621.94,3829.26L1621.94,4804.93C1621.94,5074.18 1670.48,5292.77 1730.27,5292.77L5781.98,5292.77C5841.77,5292.77 5890.31,5074.18 5890.31,4804.93L5890.31,3829.26Z" />
				</g>
			</g>
		</a>
    </g>
    <path d="M2734.76,1162.2C2734.76,902.305 2523.76,691.305 2263.86,691.305L956.714,691.305C696.82,691.305 485.82,902.305 485.82,1162.2L485.82,2103.99C485.82,2363.88 696.82,2574.88 956.714,2574.88L2263.86,2574.88C2523.76,2574.88 2734.76,2363.88 2734.76,2103.99L2734.76,1162.2Z" style="fill:rgb(235,235,235);"/>
</svg>
`, nil
					},
				},

				ui.Form[MyForm, ui.Void]{
					ID: "bla-bl",
					Init: func(void ui.Void) MyForm {
						return MyForm{
							Name: ui.TextField{
								Label:    "loldel",
								Hint:     "extra",
								Disabled: true,
							},
							Check: ui.SwitchField{
								Label: "super switch",
								Hint:  "super hint",
							},
							DueDate: ui.DateField{
								Label: "fällig am",
								Hint:  "meh",
							},
							Comment: ui.TextAreaField{
								Label: "dein kommentar",
								Hint:  "darf super lang sein",
							},
						}
					},
					Load: func(form MyForm, void ui.Void) MyForm {
						form.Check.Value = false
						form.DueDate.Value = time.Now()
						form.Comment.Value = "prosa"
						return form
					},
					Submit: ui.FormAction[MyForm, ui.Void]{
						Title: "xy",
						Receive: func(form MyForm, void ui.Void) (MyForm, ui.Action) {
							form.Check.Error = "dumm"
							form.Comment.Error = form.Comment.Value
							return form, nil
						},
					},
				},
			),
		})

		cfg.Page(ui.Page[OverParams]{
			ID:              "overview",
			Unauthenticated: false,
			Title:           "Übersicht",
			Breadcrumbs: slice.Of(
				ui.Breadcrumb{
					Title:  "A",
					Target: "/jupp",
				},
				ui.Breadcrumb{
					Title:  "B",
					Target: "/jupp",
				},
				ui.Breadcrumb{
					Title: "C",
				},
			),
			Description: "Diese Übersicht zeigt alle Stadtmusikanten an. Der Nutzer kann zudem Löschen und in die Detailansicht.",
			Navigation: slice.Of(
				ui.PageNavTarget{
					Target: "overview/42/60", // TODO fix with typesafe params? problem: package cycles in Go
					Icon:   ui.FontIcon{Name: "mdi-home"},
					Title:  "Übersicht",
				},
			),
			Children: slice.Of[ui.Component[OverParams]](
				ui.Timeline[OverParams]{
					ID: "super-time",
					Items: slice.Of(
						ui.TimelineItem{
							Title: "Register",
						},
						ui.TimelineItem{
							Title: "Name",
							Icon:  ui.FontIcon{Name: "mdi-alert"},
						},
						ui.TimelineItem{
							Title: "Adresse",
							Color: "error",
						},
						ui.TimelineItem{
							Title:  "Art",
							Target: "/jupp",
						},
					),
				},

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
							Chooser: ui.SelectField[PID]{
								Label:       "Wähle einen",
								SelectedIDs: []PID{"2"},
								Hint:        "genau etwas aus Liste",
								Multiple:    true,
								Disabled:    false,
								List: slice.Of(
									ui.SelectItem[PID]{
										ID:      "1",
										Caption: "hallo",
									},
									ui.SelectItem[PID]{
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
								Title:       "Helden" + p.User.UserID(),
								Subtitle:    "Super low code" + p.User.Email(),
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
								Title:         "a",
								Content:       ui.CardText("Toller content auf Kachel"),
								PrependIcon:   ui.FontIcon{Name: "mdi-check"},
								AppendIcon:    ui.FontIcon{Name: "mdi-account"},
								PrimaryAction: ui.Redirect{Target: "/jupp"},
							},

							ui.Card{
								Title: "Super Titel",
								Content: ui.CardMetric{
									Value: "101%",
									Icon:  ui.FontIcon{Name: "mdi-alert-circle", Color: "error"},
								},
								PrimaryAction: ui.Redirect{Target: "/jupp"},
							},

							ui.Card{
								Title:    "Helden2",
								Subtitle: "Super low code2",
								//Content:     ui.CardText("Toller content auf Kachel2"),
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
	Chooser  ui.SelectField[PID]
}
