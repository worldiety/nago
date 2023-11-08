package usecase

import (
	data2 "go.wdy.de/nago/container/data"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/example/domain/data"
)

type Ausstelleraggregat struct {
	aussteller data.AusstellerRepository
}

func (a Ausstelleraggregat) NeuenAusstellerAnlegen(name string) data.AusstellerValidation {
	a.aussteller.Save()
	return data.AusstellerValidation{
		ID: "todo",
		Vorname: data2.Validateable[string]{
			Value:     name,
			ErrorText: "not implemented",
		},
	}
}

func (a Ausstelleraggregat) AusstellerAnzeigen() slice.Slice[data.Aussteller] {
	return a.aussteller.Filter(func(aussteller data.Aussteller) bool {
		return true
	})
}

func (a Ausstelleraggregat) AusstellerLöschen(id data.Ausstellernummer) error {
	return a.aussteller.Delete(id)
}
