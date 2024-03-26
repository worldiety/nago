package main

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"go.wdy.de/nago/presentation/uix/xtable"
)

type PersonViewModel struct {
	Firstname string `caption:"Vorname"`
	Lastname  string `caption:"Nachname"`
	Age       int    `caption:"Alter"`
	Rank      Rank
}

func dataPage(wire ui.Wire, persons data.Repository[Person, PersonID]) *ui.Page {
	return ui.NewPage(wire, func(page *ui.Page) {
		page.Body().Set(ui.NewScaffold(func(scaffold *ui.Scaffold) {
			scaffold.Body().Set(xtable.NewTable(page, persons, func(e Person) PersonViewModel {
				return PersonViewModel{
					Firstname: e.Firstname,
					Lastname:  e.Lastname,
					Age:       e.Age,
					Rank:      e.Rank,
				}
			}, xtable.Options[Person, PersonID]{
				CanSearch: true,
				CanSort:   true,
				OnEdit: func(person Person) {
					xdialog.ShowMessage(page, fmt.Sprintf("Navigate to detail page of `%v`", person.ID))
				},
				OnDelete: func(person Person) error {
					return persons.DeleteByID(person.ID)
				},
			}))
		}))
	})
}
