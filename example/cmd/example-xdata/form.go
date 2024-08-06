package main

import (
	"go.wdy.de/nago/presentation/uilegacy"
	"go.wdy.de/nago/presentation/uix/xform"
)

func edit(modals uilegacy.ModalOwner, persons *PersonService, person *Person) {
	b := xform.NewBinding()
	xform.String(b, &person.Firstname, xform.Field{Label: "Vorname", Group: "Namenszeug"})
	xform.String(b, &person.Lastname, xform.Field{Label: "Nachname", Group: "Namenszeug"})
	xform.OneToMany(b, &person.Friends, persons.Persons(), PersonName, xform.Field{Label: "Freunde", Group: "Friendos"})
	xform.OneToOne(b, &person.BestFriend, persons.Persons(), PersonName, xform.Field{Label: "Bester Freund", Group: "Friendos"})
	xform.Bool(b, &person.CoolGuy, xform.Field{Label: "Knorke?"})
	xform.Date(b, &person.Birthday, xform.Field{Label: "Geburtstag"})
	xform.Int(b, &person.Age, xform.Field{Label: "Alter"})
	xform.Slider(b, &person.Rank, 200, 1000, 5, xform.Field{Label: "Rang"})

	xform.Show(modals, b, func() error {
		return persons.Update(*person)
	})
}
