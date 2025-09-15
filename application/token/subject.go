// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std/tick"
	"golang.org/x/text/language"
)

type subject struct {
	repo           Repository
	token          Token
	mutex          sync.Mutex
	roles          map[role.ID]struct{}
	groups         map[group.ID]struct{}
	licenses       map[license.ID]struct{}
	permissions    map[permission.ID]struct{}
	allPermissions []permission.ID
	lastLoaded     time.Time
	findRoleByID   role.FindByID
}

func (s *subject) HasResourcePermission(name string, id string, p permission.ID) bool {
	//TODO implement me
	panic("implement me")
}

func newSubject(findRoleByID role.FindByID, repo Repository, token Token) *subject {
	if token.Impersonation.IsSome() {
		panic(fmt.Errorf("impersonation is not allowed"))
	}

	return &subject{
		repo:         repo,
		token:        token,
		findRoleByID: findRoleByID,
		roles:        make(map[role.ID]struct{}),
		groups:       make(map[group.ID]struct{}),
		licenses:     make(map[license.ID]struct{}),
		permissions:  make(map[permission.ID]struct{}),
	}
}

func (s *subject) Audit(permission permission.ID) error {
	if !s.Valid() {
		return user.PermissionDeniedErr
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.permissions[permission]; !ok {
		return user.PermissionDeniedErr
	}

	return nil
}

func (s *subject) AuditResource(name string, id string, p permission.ID) error {
	if !s.Valid() {
		return user.PermissionDeniedErr
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	perms := s.token.Resources[user.Resource{
		Name: name,
		ID:   id,
	}]

	for _, perm := range perms {
		if perm == p {
			return nil
		}
	}

	return user.PermissionDeniedErr
}

func (s *subject) HasPermission(permission permission.ID) bool {
	s.load()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.permissions[permission]
	return ok
}

func (s *subject) Permissions() iter.Seq[permission.ID] {
	s.load()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return slices.Values(s.allPermissions)
}

func (s *subject) ID() user.ID {
	return user.ID(s.token.ID)
}

func (s *subject) Name() string {
	s.load()
	return s.token.Name
}

func (s *subject) Firstname() string {
	s.load()
	return s.token.Name
}

func (s *subject) Lastname() string {
	return ""
}

func (s *subject) Email() string {
	return ""
}

func (s *subject) Avatar() string {
	return ""
}

func (s *subject) Roles() iter.Seq[role.ID] {
	s.load()
	return slices.Values(s.token.Roles)
}

func (s *subject) HasRole(id role.ID) bool {
	s.load()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.roles[id]
	return ok
}

func (s *subject) Groups() iter.Seq[group.ID] {
	s.load()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return slices.Values(s.token.Groups)
}

func (s *subject) HasGroup(id group.ID) bool {
	s.load()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.groups[id]
	return ok
}

func (s *subject) HasLicense(id license.ID) bool {
	s.load()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.licenses[id]
	return ok
}

func (s *subject) Licenses() iter.Seq[license.ID] {
	s.load()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return slices.Values(s.token.Licenses)
}

func (s *subject) Valid() bool {
	s.load()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.token.ValidUntil.IsZero() || tick.Now(tick.Minute).Before(s.token.ValidUntil.Time(tick.Now(tick.Minute).Location()))
}

func (s *subject) Language() language.Tag {
	return language.English
}

func (s *subject) Bundle() *i18n.Bundle {
	bnd, _ := i18n.Default.MatchBundle(s.Language())
	return bnd
}

func (s *subject) load() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.lastLoaded == tick.Now(tick.Minute) {
		return
	}

	s.lastLoaded = tick.Now(tick.Minute)
	clear(s.roles)
	clear(s.groups)
	clear(s.licenses)
	clear(s.permissions)
	clear(s.allPermissions)

	optToken, err := s.repo.FindByID(s.token.ID)
	if err != nil {
		slog.Error("cannot load token", "err", err.Error())
		s.token = Token{ID: s.token.ID} // clear it entirely
		return
	}

	if optToken.IsNone() {
		slog.Error("failed to find token", "id", s.token.ID)
		s.token = Token{ID: s.token.ID} // clear it entirely
		return
	}

	s.token = optToken.Unwrap()
	for _, id := range s.token.Roles {
		s.roles[id] = struct{}{}
	}

	for _, id := range s.token.Groups {
		s.groups[id] = struct{}{}
	}

	for _, id := range s.token.Licenses {
		s.licenses[id] = struct{}{}
	}

	for _, id := range s.token.Permissions {
		s.permissions[id] = struct{}{}
	}

	for _, rid := range s.token.Roles {
		optRole, err := s.findRoleByID(user.SU(), rid)
		if err != nil {
			slog.Error("failed to find role by id", "id", rid, "err", err.Error())
			continue
		}

		if optRole.IsNone() {
			// stale ref is just gone, ignore
			continue
		}

		for _, id := range optRole.Unwrap().Permissions {
			s.permissions[id] = struct{}{}
		}
	}

	s.allPermissions = slices.Collect(maps.Keys(s.permissions))
	slices.Sort(s.allPermissions)
}
