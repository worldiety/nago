package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
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
		persons.SaveAll(func(yield func(Person) bool) {
			yield(Person{
				ID:        "1",
				Firstname: "bilbo",
				Lastname:  "beutlin",
				Score:     1,
			})

			yield(Person{
				ID:         "2",
				Firstname:  "Frodo",
				Lastname:   "Beutlin",
				BestFriend: "1",
				Score:      2,
			})
		})

		example := Person{Firstname: "Mr. Singleton"}

		cfg.Component(".", func(wnd core.Window) core.View {
			bnd := crud.NewBinding[Person](wnd, "1234")
			bnd.Add(
				crud.Text("Vorname", func(entity *Person) *string {
					return &entity.Firstname
				}).WithValidation(func(person Person) (errorText string, infrastructureError error) {
					if person.Firstname != "Torben" {
						return "Du bist nicht Torben", nil
					}

					return "", nil
				}).WithSupportingText("Gib Torben ein"),
				crud.Text("Nachname", func(entity *Person) *string {
					return &entity.Lastname
				}),
			)

			opts := &crud.Options[Person, PID]{}

			opts.FindAll(persons.Each)

			return ui.VStack(
				crud.NewView[Person, PID](wnd, opts, bnd),
				crud.Form[Person](bnd, &example),
			)
		})
	}).Run()
}
