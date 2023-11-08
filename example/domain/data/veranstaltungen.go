package data

import dm "go.wdy.de/nago/domain"

type Veranstaltungsnummer string

type Veranstaltung struct {
	ID   Veranstaltungsnummer
	Name string
	Jahr int
}

func (k Veranstaltung) Identity() Veranstaltungsnummer {
	return k.ID
}

type VeranstaltungsRepository dm.Repository[Veranstaltung, Veranstaltungsnummer]
