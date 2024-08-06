package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/crud"
	"go.wdy.de/nago/web/vuejs"
)

type PID string
type Score int
type Person struct {
	ID         PID
	Firstname  string
	Lastname   string
	Friends    []PID // this is like foreign keys, however they become stale and are not automatically updated
	BestFriend PID   // this the same as above, but the one-to-one case
	Score      Score
}

func (p Person) Identity() PID {
	return p.ID
}

type Persons data.Repository[Person, PID]

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		persons := application.SloppyRepository[Person, PID](cfg)

		cfg.Component(".", func(wnd core.Window) core.View {
			return uilegacy.NewPage(func(page *uilegacy.Page) {
				page.Body().Set(crud.NewView[Person](page, crud.NewOptions(func(opts *crud.Options[Person]) {
					opts.
						Responsive(wnd). // you need to provide a window to make the view responsive
						PrepareCreate(func(person Person) (Person, error) {
							person.ID = data.RandIdent[PID]()
							return person, nil
						}).
						Create(persons.Save).
						ReadAll(persons.Each).
						Update(persons.Save).
						Delete(persons.DeleteByEntity).
						AggregateActions(
							crud.AggregateAction[Person]{ // we also support more actions
								Icon:    icon.UserPlus,
								Caption: "Freunde zum Feiern einladen",
								Action: func(uilegacy.ModalOwner, Person) error {
									// call something in our domain
									return nil
								},
								Style: "",
								// we can also show and hide any actions based on a property of our model
							}.WithOptions(crud.AggregationActionOptionVisibility(func(p Person) bool { return len(p.Friends) > 0 })),
						).
						Bind(func(bnd *crud.Binding[Person]) {
							crud.Text(bnd,
								crud.FromPtr("ID", func(model *Person) *PID {
									return &model.ID
								}, crud.RenderHints{
									crud.Create: crud.Hidden,
									crud.Update: crud.ReadOnly,
									crud.Card:   crud.Title, // you can customize the responsive card title
								}),
							)
							crud.Text(bnd, crud.FromPtr("Vorname", func(model *Person) *string {
								return &model.Firstname
							}))
							crud.Text(bnd, crud.FromPtr("Nachname", func(model *Person) *string {
								return &model.Lastname
							}))
							crud.Int(bnd, crud.FromPtr("Score", func(model *Person) *Score {
								return &model.Score
							}))
							crud.OneToMany(bnd, persons.Each, func(person Person) string {
								return person.Firstname
							}, crud.FromPtr("Freunde", func(model *Person) *[]PID {
								return &model.Friends
							}))

							crud.OneToOne(bnd, persons.Each, func(person Person) string {
								return person.Firstname
							}, crud.FromPtr("Bester Freund", func(model *Person) *PID {
								return &model.BestFriend
							}))
						})

				})))
			})
		})
	}).Run()
}
