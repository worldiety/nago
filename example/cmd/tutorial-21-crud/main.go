package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/web/vuejs"
	"log"
	"log/slog"
	"net/http"
)

import _ "net/http/pprof"

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
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

		cfg.Component(".", func(wnd core.Window) core.View {
			bnd := crud.NewBinding[Person](wnd)
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

				crud.AggregateActions("Optionen",
					crud.ButtonDelete[Person](wnd, func(p Person) error {
						slog.Info("delete person", "id", p.ID)
						return iam.PermissionDeniedError("bla")
					}),
					crud.ButtonEdit[Person](wnd, bnd, func(p Person) (string, error) {
						slog.Info("update person", "id", p.ID, p)
						return "", persons.Save(p)
					}),
				),
			)

			return ui.VStack(
				crud.NewView[Person, PID](
					crud.Options[Person, PID](wnd, bnd).
						FindAll(persons.Each).
						Title("Personen"),
				),
				//	crud.Form[Person](bnd, &example),
				//	crud.Card[Person](bnd, &example).Frame(ui.Frame{Width: ui.L320}),
			).Padding(ui.Padding{}.All(ui.L16))
		})
	}).Run()
}
