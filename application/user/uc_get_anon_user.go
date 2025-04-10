// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xiter"
	"golang.org/x/text/language"
	"iter"
	"log/slog"
	"slices"
	"sync/atomic"
)

func NewGetAnonUser(loadGlobal settings.LoadGlobal, findRoleByID role.FindByID, bus events.Bus) GetAnonUser {
	var subj atomic.Pointer[anonSubject]
	loadSubject := func() {
		subject := createAnonSubject(loadGlobal, findRoleByID)
		subj.Store(&subject)
	}

	loadSubject()

	events.SubscribeFor[settings.GlobalSettingsUpdated](bus, func(evt settings.GlobalSettingsUpdated) {
		if _, ok := evt.Settings.(Settings); ok {
			loadSubject()
		}
	})

	events.SubscribeFor[role.Updated](bus, func(evt role.Updated) {
		loadSubject()
	})

	events.SubscribeFor[role.Deleted](bus, func(evt role.Deleted) {
		loadSubject()
	})

	return func() Subject {
		return subj.Load()
	}
}

func createAnonSubject(loadGlobal settings.LoadGlobal, findRoleByID role.FindByID) anonSubject {
	cfg := settings.ReadGlobal[Settings](loadGlobal)
	anon := anonSubject{
		groupsMap:           map[group.ID]struct{}{},
		rolesMap:            map[role.ID]struct{}{},
		permissionsMap:      map[permission.ID]struct{}{},
		resourcePermissions: map[Resource]map[permission.ID]struct{}{},
	}

	for _, anonGroup := range cfg.AnonGroups {
		anon.groupsMap[anonGroup] = struct{}{}
	}

	for _, anonRole := range cfg.AnonRoles {
		anon.rolesMap[anonRole] = struct{}{}
		optRole, err := findRoleByID(SU(), anonRole)
		if err != nil {
			slog.Error("anon: role cannot get loaded", "id", anonRole, "err", err.Error())
			continue
		}

		if optRole.IsNone() {
			// may be normal, just gone
			continue
		}

		r := optRole.Unwrap()
		for _, pid := range r.Permissions {
			anon.permissionsMap[pid] = struct{}{}
			anon.permissions = append(anon.permissions, pid)
		}
	}

	// todo resource permissions not yet implemented

	return anon
}

type anonSubject struct {
	groupsMap           map[group.ID]struct{}
	groups              []group.ID
	rolesMap            map[role.ID]struct{}
	roles               []role.ID
	permissionsMap      map[permission.ID]struct{}
	permissions         []permission.ID
	resourcePermissions map[Resource]map[permission.ID]struct{}

	// intentionally anon users never support Licenses because any amount of users share the same anon subject
	// and a per-user license would be pointless.
}

func (a anonSubject) Audit(permission permission.ID) error {
	// security note: even though we are not valid, we pass audit intentionally
	if !a.HasPermission(permission) {
		return PermissionDeniedErr
	}

	return nil
}

func (a anonSubject) HasPermission(permission permission.ID) bool {
	_, ok := a.permissionsMap[permission]
	return ok
}

func (a anonSubject) Permissions() iter.Seq[permission.ID] {
	return slices.Values(a.permissions)
}

func (a anonSubject) AuditResource(name string, id string, p permission.ID) error {
	if !a.HasResourcePermission(name, id, p) {
		return PermissionDeniedErr
	}

	return nil
}

func (a anonSubject) HasResourcePermission(name string, id string, p permission.ID) bool {
	if a.HasPermission(p) {
		return true
	}

	perms, ok := a.resourcePermissions[Resource{name, id}]
	if !ok {
		return false
	}

	if _, ok := perms[p]; ok {
		return true
	}

	return false
}

func (a anonSubject) ID() ID {
	return ""
}

func (a anonSubject) Name() string {
	return ""
}

func (a anonSubject) Firstname() string {
	return ""
}

func (a anonSubject) Lastname() string {
	return ""
}

func (a anonSubject) Email() string {
	return ""
}

func (a anonSubject) Avatar() string {
	return ""
}

func (a anonSubject) Roles() iter.Seq[role.ID] {
	return slices.Values(a.roles)
}

func (a anonSubject) HasRole(id role.ID) bool {
	_, ok := a.rolesMap[id]
	return ok
}

func (a anonSubject) Groups() iter.Seq[group.ID] {
	return slices.Values(a.groups)
}

func (a anonSubject) HasGroup(id group.ID) bool {
	_, ok := a.groupsMap[id]
	return ok
}

func (a anonSubject) HasLicense(id license.ID) bool {
	return false // always false by definition. Never change this, it is not an error.
}

func (a anonSubject) Licenses() iter.Seq[license.ID] {
	return xiter.Empty[license.ID]()
}

func (a anonSubject) Valid() bool {
	return false // security note: we are always an invalid user, however an Audit and Permission check will succeed.
}

func (a anonSubject) Language() language.Tag {
	return language.German
}
