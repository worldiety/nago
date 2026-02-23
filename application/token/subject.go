// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"sync"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std/tick"
	"golang.org/x/text/language"
)

type subject struct {
	repo         Repository
	token        Token
	mutex        sync.Mutex
	rdb          *rebac.DB
	lastLoaded   time.Time
	findRoleByID role.FindByID
	ctx          context.Context
}

func newSubject(ctx context.Context, findRoleByID role.FindByID, repo Repository, token Token, rdb *rebac.DB) *subject {
	if token.Impersonation.IsSome() {
		panic(fmt.Errorf("impersonation is not allowed"))
	}

	return &subject{
		ctx:          ctx,
		repo:         repo,
		token:        token,
		findRoleByID: findRoleByID,
		rdb:          rdb,
	}
}

func (s *subject) Context() context.Context {
	return s.ctx
}

func (s *subject) HasResourcePermission(name rebac.Namespace, id rebac.Instance, p permission.ID) bool {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Audit(perm permission.ID) error {
	if !s.Valid() {
		return user.PermissionDeniedErr
	}

	if !s.HasPermission(perm) {
		var name = string(perm)
		if p, ok := permission.Find(perm); ok {
			name = p.Name
		}

		return user.PermissionDeniedError(name)
	}

	return nil
}

func (s *subject) source() rebac.Entity {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.token.Impersonation.IsSome() {
		usr := s.token.Impersonation.Unwrap()
		return rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(usr)}
	}

	return rebac.Entity{Namespace: Namespace, Instance: rebac.Instance(s.token.ID)}
}

func (s *subject) AuditResource(name rebac.Namespace, id rebac.Instance, p permission.ID) error {
	if !s.Valid() {
		return user.PermissionDeniedErr
	}

	if !s.HasResourcePermission(name, id, p) {
		var permName = string(p)
		if perm, ok := permission.Find(p); ok {
			permName = perm.Name
		}

		return user.PermissionDeniedError(permName)
	}

	return nil
}

func (s *subject) HasPermission(permission permission.ID) bool {
	s.load()

	ok, err := s.rdb.Contains(rebac.Triple{
		Source:   s.source(),
		Relation: rebac.Relation(permission),
		Target: rebac.Entity{
			Namespace: rebac.Global,
			Instance:  rebac.AllInstances,
		},
	})

	if err != nil {
		slog.Error("cannot check resource permission", "err", err)
		return false
	}

	return ok
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

	return func(yield func(role.ID) bool) {
		it := s.rdb.Query(rebac.Select().
			Where().Source().IsNamespace(role.Namespace).
			// instance==?
			Where().Relation().Has(rebac.Member).
			Where().Target().Set(s.source()),
		)

		for triple, err := range it {
			if err != nil {
				slog.Error("cannot iterate roles", "err", err)
				return
			}

			if !yield(role.ID(triple.Target.Instance)) {
				return
			}
		}
	}
}

func (s *subject) HasRole(id role.ID) bool {
	s.load()
	ok, err := s.rdb.Contains(rebac.Triple{
		Source: rebac.Entity{
			Namespace: role.Namespace,
			Instance:  rebac.Instance(id),
		},
		Relation: rebac.Member,
		Target:   s.source(),
	})

	if err != nil {
		slog.Error("cannot check role membership", "err", err)
		return false
	}

	return ok
}

func (s *subject) Groups() iter.Seq[group.ID] {
	s.load()

	return func(yield func(group.ID) bool) {
		it := s.rdb.Query(rebac.Select().
			Where().Source().IsNamespace(group.Namespace).
			Where().Relation().Has(rebac.Member).
			Where().Target().Set(s.source()),
		)

		for triple, err := range it {
			if err != nil {
				slog.Error("cannot iterate roles", "err", err)
				return
			}

			if !yield(group.ID(triple.Target.Instance)) {
				return
			}
		}
	}
}

func (s *subject) HasGroup(id group.ID) bool {
	s.load()

	ok, err := s.rdb.Contains(rebac.Triple{
		Source: rebac.Entity{
			Namespace: group.Namespace,
			Instance:  rebac.Instance(id),
		},
		Relation: rebac.Member,
		Target:   s.source(),
	})

	if err != nil {
		slog.Error("cannot check group membership", "err", err)
		return false
	}

	return ok
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

}
