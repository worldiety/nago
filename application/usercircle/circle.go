package usercircle

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/image"
	"go.wdy.de/nago/pkg/data"
)

type ID string

type Circle struct {
	Avatar image.ID `json:"avatar,omitempty" style:"avatar"`
	ID     ID       `json:"ID,omitempty" visible:"false"`
	// Name of the circle
	Name        string `json:"name,omitempty" `
	Description string `json:"description,omitempty" label:"Beschreibung" lines:"3"`
	// Administrators are users which can add or remove any user to none or any of the defined roles.
	Administrators []user.ID `json:"administrators,omitempty" table-visible:"false" label:"Verwaltende Nutzer" source:"nago.users" supportingText:"Die hier ausgewählten Nutzer können nach freien Ermessen Rollen zu den im Kreis enthaltenen Nutzern hinzufügen oder entfernen."`
	// Roles allowed to assign to a user.
	Roles []role.ID `json:"roles,omitempty" label:"Verwaltbare Rollen" table-visible:"false" source:"nago.roles" supportingText:"Die hier ausgewählten Rollen können durch die Administratoren des Kreises hinzügefügt oder entfernt werden."`

	// Groups allowed to assign to a user.
	Groups []role.ID `json:"groups,omitempty" label:"Verwaltbare Gruppen" table-visible:"false" source:"nago.groups" supportingText:"Die hier ausgewählten Gruppen können durch die Administratoren des Kreises hinzügefügt oder entfernt werden. Achtung, wenn Mitglieder auf Basis von Gruppen ermittelt werden, kann der Administrator dieses Kreises seine Nutzer verlieren."`

	// Member Rules, if all of them are empty, all users are included in the circle.
	_                 any          `label:"Die folgenden Felder bestimmen die Regeln wie bestimmt wird, ob Nutzer zu einem Kreis gehören oder nicht. Die Regeln werden dynamisch ausgewertet und ein Nutzer kann Mitglied verschiedener Kreise gleichzeitig sein."`
	MemberRuleUsers   []user.ID    `json:"memberRuleUsers,omitempty" table-visible:"false" label:"Enthaltene Nutzer" source:"nago.users" label:"Explizite Nutzer" supportingText:"Die hier ausgewählten Nutzer sind immer festes Mitglied dieses Kreises."`
	MemberRuleDomains []user.Email `json:"memberRuleDomains,omitempty" table-visible:"false" label:"Nutzer mit EMail-Adressen" supportingText:"Alle Nutzer deren EMail mit einer dieser Domains endet. Pro Zeile wird eine Domain-Endung (z.B. @worldiety.de) ausgewertet."`
	MemberRuleGroups  []group.ID   `json:"memberRuleGroups,omitempty" table-visible:"false" label:"Nutzer mit Gruppenzugehörigkeit" source:"nago.groups"`
}

func (c Circle) String() string {
	return c.Name
}

func (c Circle) Identity() ID {
	return c.ID
}

func (c Circle) WithIdentity(id ID) Circle {
	c.ID = id
	return c
}

type Repository data.Repository[Circle, ID]
