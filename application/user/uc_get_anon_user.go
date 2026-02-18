// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"slices"
	"sync/atomic"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xslices"
	"golang.org/x/text/language"
)

func NewGetAnonUser(ctx context.Context, bus events.Bus, loadGlobal settings.LoadGlobal, findRoleByID role.FindByID, listPerms role.ListPermissions) GetAnonUser {
	var subj atomic.Pointer[anonSubject]
	loadSubject := func() {
		subject := createAnonSubject(ctx, language.English, loadGlobal, findRoleByID, listPerms)
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
		return *subj.Load()
	}
}

func createAnonSubject(ctx context.Context, tag language.Tag, loadGlobal settings.LoadGlobal, findRoleByID role.FindByID, listPerms role.ListPermissions) anonSubject {
	bnd, ok := i18n.Default.MatchBundle(tag)
	if !ok {
		panic(fmt.Errorf("unreachable"))
	}

	cfg := settings.ReadGlobal[Settings](loadGlobal)
	anon := anonSubject{
		groupsMap:      map[group.ID]struct{}{},
		rolesMap:       map[role.ID]struct{}{},
		permissionsMap: map[permission.ID]struct{}{},
		tag:            tag,
		bundle:         bnd,
		ctx:            ctx,
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
		perms, err := xslices.Collect2(listPerms(SU(), r.ID))
		if err != nil {
			slog.Error("anon: perms for role cannot get loaded", "id", anonRole, "err", err.Error())
			continue
		}

		for _, pid := range perms {
			anon.permissionsMap[pid] = struct{}{}
			anon.permissions = append(anon.permissions, pid)
		}
	}

	// todo resource permissions not yet implemented

	return anon
}

var _ Subject = anonSubject{}

type anonSubject struct {
	groupsMap      map[group.ID]struct{}
	groups         []group.ID
	rolesMap       map[role.ID]struct{}
	roles          []role.ID
	permissionsMap map[permission.ID]struct{}
	permissions    []permission.ID
	bundle         *i18n.Bundle
	tag            language.Tag
	ctx            context.Context

	// intentionally anon users never support Licenses because any amount of users share the same anon subject
	// and a per-user license would be pointless.
}

func (a anonSubject) Bundle() *i18n.Bundle {
	return a.bundle
}

func (a anonSubject) Audit(permission permission.ID) error {
	// security note: even though we are not valid, we pass audit intentionally
	if !a.HasPermission(permission) {
		return PermissionDeniedErr
	}

	return nil
}

func (a anonSubject) Context() context.Context {
	return a.ctx
}

func (a anonSubject) HasPermission(permission permission.ID) bool {
	_, ok := a.permissionsMap[permission]
	return ok
}

func (a anonSubject) AuditResource(resourceType rebac.Namespace, instance rebac.Instance, perm permission.ID) error {
	if !a.HasResourcePermission(resourceType, instance, perm) {
		return PermissionDeniedErr
	}

	return nil
}

func (a anonSubject) HasResourcePermission(resourceType rebac.Namespace, instance rebac.Instance, p permission.ID) bool {
	if a.HasPermission(p) {
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

func (a anonSubject) Valid() bool {
	return false // security note: we are always an invalid user, however an Audit and Permission check will succeed.
}

func (a anonSubject) Language() language.Tag {
	return a.tag
}
