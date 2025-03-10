package user

import (
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/settings"
)

var _ = enum.Variant[settings.GlobalSettings, Settings](
	enum.Rename[Settings]("nago.user.settings"),
)

type Settings struct {
	_                 any  `title:"Nutzerverwaltung" description:"Allgemeine Vorgaben bezüglich der Nutzerverwaltung vornehmen."`
	SelfRegistration  bool `json:"selfRegistration" label:"Freie Registrierung" supportingText:"Wenn erlaubt, dann kann sich jeder anonyme Besucher ein eigenes Konto erstellen. Ansonsten müssen die Nutzerkonten manuell durch einen Administrator erstellt werden."`
	SelfPasswordReset bool `json:"selfPasswordReset" label:"Passwort vergessen Funktion" supportingText:"Nutzer können im Self-Service ihre Kennwörter zurücksetzen. Hierfür ist ein Mail-Server erforderlich."`

	__ any `label:"Die folgenden Einschränkungen gelten für die freie Registrierung und Nutzerkreisverwaltung."`

	AllowedDomains []string   `json:"allowedDomains" lines:"5" label:"Erlaubte Domains" supportingText:"Jede Zeile stellt einen erlaubten Domänen Suffix dar, also z.B. @worldiety.de. Wenn diese Liste leer ist, darf sich jeder registrieren."`
	DefaultRoles   []role.ID  `json:"defaultRoles" source:"nago.groups" label:"Standardrolle" supportingText:"Diese Rollen werden pauschal einem neuen Nutzer hinzugefügt."`
	DefaultGroups  []group.ID `json:"defaultGroups" source:"nago.roles" label:"Standardgruppen" supportingText:"Diese Gruppen werden pauschal einem neuen Nutzer hinzugefügt."`
}

func (s Settings) GlobalSettings() bool { return true }
