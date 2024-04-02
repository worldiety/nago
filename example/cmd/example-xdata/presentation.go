package main

import (
	"fmt"
	dm "go.wdy.de/nago/domain"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"go.wdy.de/nago/presentation/uix/xtable"
)

func dataPage(wire ui.Wire, persons *PersonService) *ui.Page {
	return ui.NewPage(wire, func(page *ui.Page) {
		page.Body().Set(ui.NewScaffold(func(scaffold *ui.Scaffold) {
			scaffold.Body().Set(xtable.NewTable(page, persons.ViewPersons(), xtable.NewModelBinding[PersonView](), xtable.Options[PersonView]{
				CanSearch: true,
				AggregateActions: []xtable.AggregateAction[PersonView]{
					xtable.NewEditAction(func(pview PersonView) error {
						person, err := std.Unpack2(persons.FindPerson(pview.ID))
						if err != nil {
							return err
						}

						edit(page, persons, &person)
						return nil
					}),
					xtable.NewDeleteAction[PersonView](persons.RemoveByPersonView),
					{
						Icon:    icon.Cog6Tooth,
						Caption: "Einstellungen",
						Action: func(person PersonView) error {
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
