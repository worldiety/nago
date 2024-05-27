package iam

import (
	"go.wdy.de/nago/pkg/maps"
	"go.wdy.de/nago/pkg/slices"
	slices2 "slices"
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
		iamPerm{
			id:   ReadPermission,
			name: "Berechtigungen anzeigen",
			desc: "Träger dieser Berechtigung können alle vorhandenen Berechtigungen inkl. der Erläuterungstexte einsehen. Die Menge der Berechtigungen wird vom System vorgegeben und kann nicht dynamisch geändert werden.",
		},
		iamPerm{
			id:   ReadRole,
			name: "Rollen anzeigen",
			desc: "Träger dieser Berechtigung können alle vorhandenen Rollen und die ihnen zugeordneten Berechtigungen anzeigen.",
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
	slices2.Sort(sorted)
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
