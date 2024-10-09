package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/presentation/ui/timepicker"
	"go.wdy.de/nago/web/vuejs"
	"log"
	"log/slog"
	"net/http"
)

import _ "net/http/pprof"

type PID string
type Score int
type Grade float64
type Color string
type Birthday xtime.Date
type Vacation struct {
	Start xtime.Date
	End   xtime.Date
}

type TimeInSec int64

type Person struct {
	ID            PID
	Title         string
	Firstname     string
	Lastname      string
	Friends       []PID           // this is like foreign keys, however they become stale and are not automatically updated
	BestFriend    std.Option[PID] // this the same as above, but the one-to-one case
	Score         Score
	Grade         Grade
	Proofed       bool
	FavoriteColor std.Option[Color]
	Colors        []Color
	Birthday      Birthday
	Vacation      Vacation
	WorkDuration  TimeInSec
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
				BestFriend: std.Some[PID]("1"),
				Score:      2,
			})
		})

		cfg.RootView(".", func(wnd core.Window) core.View {
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

				crud.Int("Score", func(model *Person) *Score {
					return &model.Score
				}),

				crud.Float("Grade", func(model *Person) *Grade {
					return &model.Grade
				}),

				crud.Bool("Proofed", func(model *Person) *bool {
					return &model.Proofed
				}).WithSupportingText("Check me if this is ok"),

				crud.BoolToggle("Proofed2", func(model *Person) *bool {
					return &model.Proofed
				}).WithSupportingText("don't bind same fields into the same form"),

				crud.PickMultiple("Colors", []Color{"red", "green", "blue"}, func(model *Person) *[]Color {
					return &model.Colors
				}),

				crud.PickOne("Color", []Color{"red", "green", "blue"}, func(model *Person) *std.Option[Color] {
					return &model.FavoriteColor
				}),

				crud.OneToMany("Friendos", persons.All(), func(t Person) core.View {
					return ui.Text(fmt.Sprintf("%s %s", t.Firstname, t.Lastname))
				}, func(model *Person) *[]PID {
					return &model.Friends
				}),

				crud.OneToOne("Best Friend", persons.All(), func(t Person) core.View {
					return ui.Text(fmt.Sprintf("%s %s", t.Firstname, t.Lastname))
				}, func(model *Person) *std.Option[PID] {
					return &model.BestFriend
				}),

				crud.Date("Geburtstag", func(model *Person) *Birthday {
					return &model.Birthday
				}),

				crud.DateRange("Urlaub", func(model *Person) *xtime.Date {
					return &model.Vacation.Start
				}, func(model *Person) *xtime.Date {
					return &model.Vacation.End
				}),

				crud.Time("Arbeitszeit", 1, false, true, true, false, timepicker.DecomposedFormat, func(model *Person) *TimeInSec {
					return &model.WorkDuration
				}),

				crud.AggregateActions("Optionen",
					crud.Optional[Person](crud.ButtonDelete[Person](wnd, func(p Person) error {
						slog.Info("delete person", "id", p.ID)
						return iam.PermissionDeniedError("bla")
					}), func(person Person) bool {
						return person.ID != "1"
					}),

					crud.ButtonEdit[Person](bnd, func(p Person) (string, error) {
						slog.Info("update person", "id", p.ID, p)
						return "", persons.Save(p)
					}),
				),
			)

			return ui.VStack(
				crud.View[Person, PID](
					crud.Options[Person, PID](bnd).
						Actions(
							crud.ButtonCreate[Person](bnd, Person{ID: data.RandIdent[PID]()}, func(person Person) (errorText string, infrastructureError error) {
								if person.Firstname == "" {
									return "Vorname darf nicht leer sein", nil
								}

								return "", persons.Save(person)
							}),
						).
						FindAll(persons.All()).
						Title("Personen"),
				),
			).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.FullWidth())
		})
	}).Run()
}
