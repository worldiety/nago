package usercircle

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"slices"
	"strings"
)

type ID string

type Circle struct {
	Avatar image.ID `json:"avatar,omitempty" style:"avatar" label:""`
	ID     ID       `json:"ID,omitempty" visible:"false"`
	// Name of the circle
	Name        string `json:"name,omitempty" `
	Description string `json:"description,omitempty" label:"Beschreibung" lines:"3"`
	// Administrators are users which can add or remove any user to none or any of the defined roles.
	Administrators []user.ID `json:"administrators,omitempty" table-visible:"false" label:"Verwaltende Nutzer" source:"nago.users" supportingText:"Die hier ausgewählten Nutzer können nach freien Ermessen Rollen zu den im Kreis enthaltenen Nutzern hinzufügen oder entfernen."`
	// Roles allowed to assign to a user.
	Roles []role.ID `json:"roles,omitempty" label:"Verwaltbare Rollen" table-visible:"false" source:"nago.roles" supportingText:"Die hier ausgewählten Rollen können durch die Administratoren des Kreises hinzugefügt oder entfernt werden. Sind keine Rollen ausgewählt, ist die Rollenverwaltung nicht verfügbar."`

	// Groups allowed to assign to a user.
	Groups []group.ID `json:"groups,omitempty" label:"Verwaltbare Gruppen" table-visible:"false" source:"nago.groups" supportingText:"Die hier ausgewählten Gruppen können durch die Administratoren des Kreises hinzugefügt oder entfernt werden. Achtung, wenn Mitglieder auf Basis von Gruppen ermittelt werden, kann der Administrator dieses Kreises seine Nutzer verlieren. Sind keine Gruppen ausgewählt, ist die Gruppenverwaltung nicht verfügbar."`

	Licenses []license.ID `json:"licenses,omitempty" label:"Verwaltbare Lizenzen" table-visible:"false" source:"nago.licenses.user" supportingText:"Die hier ausgewählten Lizenzen können durch die Administratoren des Kreises hinzugefügt oder entfernt werden. Sind keine Lizenzen ausgewählt, ist die Lizenzverwaltung nicht verfügbar."`

	CanDelete  bool `json:"canDelete" table-visible:"false" label:"Nutzer löschen" supportingText:"Administratoren dürfen Nutzer aus dem System unwiderruflich entfernen."`
	CanDisable bool `json:"canDisable" table-visible:"false" label:"Nutzer deaktivieren" supportingText:"Administratoren dürfen Nutzer im System deaktivieren."`
	CanEnable  bool `json:"canEnable" table-visible:"false" label:"Nutzer aktivieren" supportingText:"Administratoren dürfen Nutzer im System aktivieren."`
	CanVerify  bool `json:"canVerify" table-visible:"false" label:"Nutzer verifizieren" supportingText:"Administratoren dürfen Nutzer im System als verifiziert markieren, obwohl diese ihre EMail-Adresse nie selbst bestätigt haben."`

	// Member Rules, if all of them are empty, all users are included in the circle.
	_                        any          `section:"Mitgliedschaft" label:"Die folgenden Felder bestimmen die Regeln wie bestimmt wird, ob Nutzer zu einem Kreis gehören oder nicht. Die Regeln werden dynamisch ausgewertet und ein Nutzer kann Mitglied verschiedener Kreise gleichzeitig sein."`
	MemberRuleUsers          []user.ID    `section:"Mitgliedschaft" json:"memberRuleUsers,omitempty" table-visible:"false" label:"Enthaltene Nutzer" source:"nago.users" label:"Explizite Nutzer" supportingText:"Die hier ausgewählten Nutzer sind immer festes Mitglied dieses Kreises."`
	MemberRuleDomains        []user.Email `section:"Mitgliedschaft" json:"memberRuleDomains,omitempty" table-visible:"false" label:"Nutzer mit E-Mail-Adressen" supportingText:"Alle Nutzer deren E-Mail mit einer dieser Domains endet. Pro Zeile wird eine Domain-Endung (z.B. @worldiety.de) ausgewertet."`
	MemberRuleRoles          []role.ID    `section:"Mitgliedschaft" json:"memberRuleRoles,omitempty" table-visible:"false" label:"Nutzer mit Rollenzugehörigkeit" source:"nago.roles"`
	MemberRuleGroups         []group.ID   `section:"Mitgliedschaft" json:"memberRuleGroups,omitempty" table-visible:"false" label:"Nutzer mit Gruppenzugehörigkeit" source:"nago.groups"`
	MemberRuleUsersBlacklist []user.ID    `section:"Mitgliedschaft" label:"Nicht enthaltene Nutzer" table-visible:"false" source:"nago.users" label:"Explizite Nicht-Nutzer" supportingText:"Die hier ausgewählten Nutzer sind in jedem Fall niemals Mitglied dieses Kreises."`
}

// isMember is a quite slow implementation. If you need to be faster, try [MyCircleMembers].
func (c Circle) isMember(usr user.User) bool {
	if slices.Contains(c.MemberRuleUsersBlacklist, usr.ID) {
		return false
	}

	if slices.Contains(c.MemberRuleUsers, usr.ID) {
		return true
	}

	for _, id := range usr.Groups {
		if slices.Contains(c.MemberRuleGroups, id) {
			return true
		}
	}

	for _, id := range usr.Roles {
		if slices.Contains(c.MemberRuleRoles, id) {
			return true
		}
	}

	for _, domain := range c.MemberRuleDomains {
		if strings.HasSuffix(string(usr.Email), string(domain)) {
			return true
		}
	}

	return false
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
