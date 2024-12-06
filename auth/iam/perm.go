package iam

import (
	"maps"
	"slices"
)

const (
	CreateUser = "de.worldiety.ora.user.create"
	ReadUser   = "de.worldiety.ora.user.read"
	UpdateUser = "de.worldiety.ora.user.update"
	DeleteUser = "de.worldiety.ora.user.delete"

	ReadPermission = "de.worldiety.ora.permission.read"

	ReadRole   = "de.worldiety.ora.role.read"
	CreateRole = "de.worldiety.ora.role.create"
	UpdateRole = "de.worldiety.ora.role.update"
	DeleteRole = "de.worldiety.ora.role.delete"

	ReadGroup   = "de.worldiety.ora.group.read"
	CreateGroup = "de.worldiety.ora.group.create"
	UpdateGroup = "de.worldiety.ora.group.update"
	DeleteGroup = "de.worldiety.ora.group.delete"
)

type iamPerm struct {
	id   string
	name string
	desc string
}

func (b iamPerm) Identity() string {
	return b.id
}

func (b iamPerm) Name() string {
	return b.name
}

func (b iamPerm) Desc() string {
	return b.desc
}

func BuildInPermissions() []Permission {
	return []Permission{
		// users
		iamPerm{
			id:   CreateUser,
			name: "Nutzer anlegen",
			desc: "Träger dieser Berechtigung können neue Nutzer anlegen.",
		},
		iamPerm{
			id:   ReadUser,
			name: "Nutzer anzeigen und auflisten",
			desc: "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen und auflisten. Das Passwort ist nicht gespeichert und kann technisch nicht eingesehen werden.",
		},
		iamPerm{
			id:   UpdateUser,
			name: "Nutzer aktualisieren",
			desc: "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften aktualisieren. Dazu gehört u.a. das Aktivieren und Deaktivieren von Nutzern, aber auch das Setzen eines neuen Kennwortes.",
		},
		iamPerm{
			id:   DeleteUser,
			name: "Nutzer löschen",
			desc: "Träger dieser Berechtigung können vorhandene Nutzer löschen.",
		},
		// permission are hardcoded and it does not make sense that these are in any way dynamic
		iamPerm{
			id:   ReadPermission,
			name: "Berechtigungen anzeigen",
			desc: "Träger dieser Berechtigung können alle vorhandenen Berechtigungen inkl. der Erläuterungstexte einsehen. Die Menge der Berechtigungen wird vom Sys vorgegeben und kann nicht dynamisch geändert werden.",
		},
		// roles
		iamPerm{
			id:   ReadRole,
			name: "Rollen anzeigen",
			desc: "Träger dieser Berechtigung können alle vorhandenen Rollen und die ihnen zugeordneten Berechtigungen anzeigen.",
		},

		iamPerm{
			id:   CreateRole,
			name: "Rollen erstellen",
			desc: "Träger dieser Berechtigung können neue Rollen erstellen.",
		},

		iamPerm{
			id:   UpdateRole,
			name: "Rollen aktualisieren",
			desc: "Träger dieser Berechtigung können vorhandene Rollen ändern.",
		},

		iamPerm{
			id:   DeleteRole,
			name: "Rollen löschen",
			desc: "Träger dieser Berechtigung können Rollen entfernen.",
		},

		// groups
		iamPerm{
			id:   ReadGroup,
			name: "Gruppen anzeigen",
			desc: "Träger dieser Berechtigung können alle vorhandenen Gruppen und die ihnen zugeordneten Nutzer anzeigen.",
		},

		iamPerm{
			id:   CreateGroup,
			name: "Gruppen erstellen",
			desc: "Träger dieser Berechtigung können neue Gruppen erstellen.",
		},

		iamPerm{
			id:   UpdateGroup,
			name: "Gruppen aktualisieren",
			desc: "Träger dieser Berechtigung können vorhandene Gruppen ändern.",
		},

		iamPerm{
			id:   DeleteGroup,
			name: "Gruppen löschen",
			desc: "Träger dieser Berechtigung können Gruppen entfernen.",
		},
	}
}

// Permission is the basic contract for the permissions repository, which is used by higher level implementations.
type Permission interface {
	Identity() string
	Name() string
	Desc() string
}

type Permissions struct {
	permissions map[string]Permission
}

func PermissionsFrom[T Permission](slice []T) *Permissions {
	p := &Permissions{
		permissions: make(map[string]Permission),
	}

	// always ensure that our permissions are available.
	// However, allow that developers permission may override the description texts.
	for _, permission := range BuildInPermissions() {
		p.permissions[permission.Identity()] = permission
	}

	for _, t := range slice {
		p.permissions[t.Identity()] = t
	}

	return p
}

func (p *Permissions) Each(yield func(permission Permission, err error) bool) {
	sorted := slices.Collect(maps.Keys(p.permissions))
	slices.Sort(sorted)
	for _, t := range sorted {
		if !yield(p.permissions[t], nil) {
			return
		}
	}
}

func (p *Permissions) Has(permission string) bool {
	_, ok := p.permissions[permission]
	return ok
}

func (p *Permissions) Get(permission string) (Permission, bool) {
	perm, ok := p.permissions[permission]
	return perm, ok
}
