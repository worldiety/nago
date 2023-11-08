package data

import (
	"go.wdy.de/nago/container/data"
	dm "go.wdy.de/nago/domain"
)

type Ausstellernummer string

type Aussteller struct {
	ID      Ausstellernummer
	Vorname string
}

type AusstellerValidation struct {
	ID      Ausstellernummer
	Vorname data.Validateable[string]
}

func (k Aussteller) Identity() Ausstellernummer {
	return k.ID
}

type AusstellerRepository dm.Repository[Aussteller, Ausstellernummer]
