package main

import (
	"fmt"
	dm "go.wdy.de/nago/domain"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
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
				AggregateActions: []xtable.AggregateAction[Person]{
					xtable.NewEditAction(func(person Person) error {
						edit(page, persons, &person)
						return nil
					}),
					xtable.NewDeleteAction(func(person Person) error {
						return persons.DeleteByID(person.ID)
					}),
					{
						Icon:    icon.Cog6Tooth,
						Caption: "Einstellungen",
						Action: func(person Person) error {
							xdialog.ShowMessage(page, fmt.Sprintf("Einstellungen von %v", person.ID))
							return nil
						},
					},
				},

				Actions: []ui.LiveComponent{
					ui.NewButton(func(btn *ui.Button) {
						btn.PreIcon().Set(icon.Plus)
						btn.Caption().Set("Neu")
						btn.Action().Set(func() {
							var person Person
							person.ID = PersonID(dm.NewID())
							person.Firstname = "Nobody"

							edit(page, persons, &person)
						})
					}),
				},
			}))
		}))
	})
}
