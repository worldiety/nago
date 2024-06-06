package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/crud"
	"go.wdy.de/nago/web/vuejs"
)

type PID string
type Person struct {
	ID        PID
	Firstname string
	Lastname  string
	Friends   []PID // this is like foreign keys, however they become stale and are not automatically updated
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

		cfg.Component(".", func(wnd core.Window) core.Component {
			return ui.NewPage(func(page *ui.Page) {
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
							crud.OneToMany(bnd, persons.Each, func(person Person) string {
								return person.Firstname
							}, crud.FromPtr("Freunde", func(model *Person) *[]PID {
								return &model.Friends
							}))
						})

				})))
			})
		})
	}).Run()
}
