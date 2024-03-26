package main

import (
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xtable"
)

type PersonViewModel struct {
	Firstname string `caption:"Vorname"`
	Lastname  string `caption:"Nachname"`
	Age       int    `caption:"Alter"`
	Rank      Rank
	Friends   int `caption:"Anzahl Freunde"`
}

func dataPage(wire ui.Wire, persons Persons) *ui.Page {
	return ui.NewPage(wire, func(page *ui.Page) {
		page.Body().Set(ui.NewScaffold(func(scaffold *ui.Scaffold) {
			scaffold.Body().Set(xtable.NewTable(page, persons, func(e Person) PersonViewModel {
				return PersonViewModel{
					Firstname: e.Firstname,
					Lastname:  e.Lastname,
					Age:       e.Age,
					Rank:      e.Rank,
					Friends:   len(e.Friends),
				}
			}, xtable.Options[Person, PersonID]{
				CanSearch: true,
				CanSort:   true,
				OnEdit: func(person Person) {
					edit(page, persons, &person)
				},
				OnDelete: func(person Person) error {
					return persons.DeleteByID(person.ID)
				},
			}))
		}))
	})
}
