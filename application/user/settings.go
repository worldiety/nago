package user

import (
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/settings"
)

var _ = enum.Variant[settings.GlobalSettings, Settings]()

type Settings struct {
	_                  any  `title:"Nutzerverwaltung" description:"Allgemeine Vorgaben bezüglich der Nutzerverwaltung vornehmen."`
	PublicRegistration bool `json:"publicRegistration" label:"Öffentliche Registrierung" supportingText:"Wenn erlaubt, dann kann sich jeder anonyme Besucher ein eigenes Konto erstellen. Ansonsten müssen die Nutzerkonten manuell durch einen Administrator erstellt werden."`

	AllowedDomains []string `json:"allowedDomains" lines:"5" label:"Erlaubte Domains" supportingText:"Jede Zeile stellt einen erlaubten Domänen Suffix dar, also z.B. @worldiety.de. Wenn diese Liste leer ist, darf sich jeder registrieren."`
}

func (s Settings) GlobalSettings() bool { return true }
