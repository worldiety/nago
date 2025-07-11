// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/web/vuejs"
	"log"
	"log/slog"
	"net/http"
	"time"
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

type Person struct {
	ID             PID
	Title          string
	Firstname      string
	Lastname       string
	Friends        []PID           // this is like foreign keys, however they become stale and are not automatically updated
	BestFriend     std.Option[PID] // this the same as above, but the one-to-one case
	Score          Score
	Grade          Grade
	Proofed        bool
	FavoriteColor  std.Option[Color]
	FavoriteColor2 std.Option[ui.Color]
	Colors         []Color
	Birthday       Birthday
	Vacation       Vacation
	WorkDuration   time.Duration
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
				crud.Text(crud.TextOptions{Label: "Vorname"}, crud.Ptr(func(entity *Person) *string {
					return &entity.Firstname
				})).WithValidation(func(person Person) (errorText string, infrastructureError error) {
					if person.Firstname != "Torben" {
						return "Du bist nicht Torben", nil
					}

					return "", nil
				}).WithSupportingText("Gib Torben ein"),
				crud.Text(crud.TextOptions{Label: "Nachname"}, crud.Ptr(func(entity *Person) *string {
					return &entity.Lastname
				})),

				crud.Int(crud.IntOptions{Label: "Score"}, crud.Ptr(func(model *Person) *Score {
					return &model.Score
				})).WithValidation(func(person Person) (errorText string, infrastructureError error) {
					if person.Score != 42 {
						return "muss 42 sein", nil
					}

					return "", nil
				}),

				crud.Float(crud.FloatOptions{Label: "Grade"}, crud.Ptr(func(model *Person) *Grade {
					return &model.Grade
				})),

				crud.Bool(crud.BoolOptions{Label: "Proofed"}, crud.Ptr(func(model *Person) *bool {
					return &model.Proofed
				})).WithSupportingText("Check me if this is ok"),

				crud.BoolToggle(crud.BoolOptions{Label: "Proofed2"}, crud.Ptr(func(model *Person) *bool {
					return &model.Proofed
				})).WithSupportingText("don't bind same fields into the same form"),

				crud.PickMultiple(crud.PickMultipleOptions[Color]{Label: "Colors", Values: []Color{"red", "green", "blue"}}, crud.Ptr(func(model *Person) *[]Color {
					return &model.Colors
				})),

				crud.PickOne(crud.PickOneOptions[Color]{Label: "Color", Values: []Color{"red", "green", "blue"}}, crud.Ptr(func(model *Person) *std.Option[Color] {
					return &model.FavoriteColor
				})),

				crud.OneToMany(crud.OneToManyOptions[Person, PID]{
					Label:           "Friendos",
					ForeignEntities: persons.All(),
					ForeignPickerRenderer: func(t Person) core.View {
						return ui.Text(fmt.Sprintf("%s %s", t.Firstname, t.Lastname))
					},
				}, crud.Ptr(func(model *Person) *[]PID {
					return &model.Friends
				})),

				crud.OneToOne(crud.OneToOneOptions[Person, PID]{
					Label:           "Best Friend",
					ForeignEntities: persons.All(),
					ForeignPickerRenderer: func(t Person) core.View {
						return ui.Text(fmt.Sprintf("%s %s", t.Firstname, t.Lastname))
					},
				}, crud.PropertyFuncs[Person, std.Option[PID]](func(p *Person) std.Option[PID] {
					return p.BestFriend
				}, func(dst *Person, v std.Option[PID]) {
					dst.BestFriend = v
				})),

				crud.Date(crud.DateOptions{Label: "Geburtstag"}, crud.Ptr(func(model *Person) *Birthday {
					return &model.Birthday
				})),

				crud.DateRange(crud.DateRangeOptions{Label: "Urlaub"}, crud.Ptr(func(model *Person) *xtime.Date {
					return &model.Vacation.Start
				}), crud.Ptr(func(model *Person) *xtime.Date {
					return &model.Vacation.End
				})),

				crud.Time(crud.TimeOptions{Label: "Arbeitszeit", ShowHours: true, ShowSeconds: true}, crud.Ptr(func(model *Person) *time.Duration {
					return &model.WorkDuration
				})),

				crud.PickOneColor(crud.PickOneColorOptions{Label: "Lieblingsfarbe 2"}, crud.Ptr(func(model *Person) *std.Option[ui.Color] {
					return &model.FavoriteColor2
				})),

				crud.AggregateActions("Optionen",
					crud.Optional[Person](crud.ButtonDelete[Person](wnd, func(p Person) error {
						slog.Info("delete person", "id", p.ID)
						return nil
					}), func(person Person) bool {
						return person.ID != "1"
					}),

					crud.ButtonEdit[Person](bnd, func(p Person) (string, error) {
						slog.Info("update person", "id", p.ID, "person", p)
						return "", persons.Save(p)
					}),
				),
			)

			return ui.VStack(
				crud.View[Person, PID](
					crud.Options[Person, PID](bnd).
						Actions(
							crud.ButtonCreate[Person](bnd, Person{ID: "do not randomize here"}, func(person Person) (errorText string, infrastructureError error) {
								if !bnd.Validates(person) {
									return "irgendein validation fehler, gugg hin", nil
								}

								if person.Firstname == "" {
									return "Vorname darf nicht leer sein", nil
								}

								person.ID = data.RandIdent[PID]() // create a unique ID here

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
