// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"context"
	"iter"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/tick"
	"golang.org/x/text/language"
)

type AuditableUser interface {
	permission.Auditable
	ID() ID
	Valid() bool
}

// Subject is a common contract for an authenticated identity, actor or subject.
// Different implementations may provide additional interfaces or
// expose concrete types behind it.
type Subject interface {
	permission.Auditable

	// AuditResource is like Audit, but instead of using the user permissions, the associated resource table
	// is also evaluated. The order of evaluation works as follows:
	//  - first check, if the given permission has been assigned to the user globally
	//  - second check, if the given id is empty and the name and permission matches, the audit is successful
	//  - third check, if the all 3 properties match, the audit is successful
	//
	// In all other cases, the audit will fail. The name declares the namespace of in which the given id is naturally
	// unique, e.g. the store or repository name.
	AuditResource(resourceType rebac.Namespace, instance rebac.Instance, perm permission.ID) error

	// HasResourcePermission is like [HasResource] checks, but instead of using the user permissions, the associated
	// resource table is also evaluated. A regular use case
	// should use the [AuditResource]. However, this may be used e.g. by the UI to show or hide specific aspects.
	HasResourcePermission(resourceType rebac.Namespace, instance rebac.Instance, perm permission.ID) bool

	// ID is the unique actor id within a single NAGO instance. These IDs are generated in a secure way,
	// however, you must not expose that into the public or use it as a source of anonymization.
	// This ID will never change throughout the lifetime of the user and this instance.
	ID() ID

	// Name contains an arbitrary non-unique calling name of the identity.
	// May be something anonymous.
	// Depending on the provider, this may be even empty e.g. due to GDPR requirements.
	Name() string

	// Firstname contains the first name, if available.
	// May be something anonymous.
	// Depending on the provider, this may be even empty e.g. due to GDPR requirements.
	Firstname() string

	// Lastname contains the last name, if available.
	// May be something anonymous.
	// Depending on the provider, this may be even empty e.g. due to GDPR requirements.
	Lastname() string

	// Email contains the mail address, if available.
	// May be something anonymous.
	// Depending on the provider, this may be even empty e.g. due to GDPR requirements.
	// You probably should NEVER rely on this to verify that two identities or subjects are the same,
	// especially if the address has never been verified by a second factor (e.g. double opt-in or similar).
	// This is a string, because it remembers you, that at no time this returned value means that the mail
	// is valid in any way. Even if it has been verified once, the domain may have been deleted, or the mailbox is
	// full or locked or even worse, has been captured by a malicious party and compromised.
	Email() string

	// Avatar returns optionally a resource representing the avatar image. This may be an url, an uri or
	// any id. By default, Nago returns an [image.ID].
	Avatar() string

	// Roles yields over all associated roles. This is important if the domain needs to model
	// resource based access using role identifiers.
	Roles() iter.Seq[role.ID]

	// HasRole returns true, if the user has the associated role.
	HasRole(id role.ID) bool

	// Groups yields over all associated groups. This is important if the domain needs to model
	// resource based access using group identifiers.
	Groups() iter.Seq[group.ID]

	// HasGroup returns true, if the user is in the associated group.
	HasGroup(id group.ID) bool

	// Valid tells us, if the subject has been authenticated and potentially contains permissions.
	// If the mail has never been verified, a user will not be valid.
	Valid() bool

	// Language returns the BCP47 language tag, which encodes a language and locale.
	Language() language.Tag

	// Bundle returns the associated and localized resource bundle.
	Bundle() *i18n.Bundle

	// Context returns the current context of the subject. This may be bound to the lifecycle of the window
	// or unbounded e.g. for [SU] or [GetAnonUser].
	Context() context.Context
}

type viewImpl struct {
	user            User
	ctx             context.Context
	mutex           sync.Mutex
	repo            Repository
	lastRefreshedAt time.Time
	refreshInterval time.Duration
	locale          language.Tag
	bundle          atomic.Pointer[i18n.Bundle]
	rdb             *rebac.DB
}

func newViewImpl(ctx context.Context, rdb *rebac.DB, repo Repository, user User) *viewImpl {
	v := &viewImpl{
		ctx:             ctx,
		user:            user,
		lastRefreshedAt: time.Now(),
		refreshInterval: 5 * time.Minute,
		repo:            repo,
		rdb:             rdb,
	}

	v.load()

	return v
}

func (v *viewImpl) Context() context.Context {
	return v.ctx
}

func (v *viewImpl) invalidate() {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.lastRefreshedAt = time.Time{}
}

func (v *viewImpl) refresh() User {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.refreshInterval == 0 {
		v.refreshInterval = 5 * time.Minute
	}

	now := tick.Now(tick.Minute)
	if now.Sub(v.lastRefreshedAt) >= v.refreshInterval {
		v.load()
	}

	return v.user
}

func (v *viewImpl) load() {
	v.lastRefreshedAt = tick.Now(tick.Minute)

	if v.user.ID == "" {
		slog.Error("user has no id")
		return
	}

	optUsr, err := v.repo.FindByID(v.user.ID)
	if err != nil {
		slog.Error("cannot refresh user", "id", v.user.ID, "err", err)
		v.user = User{ID: v.user.ID, Status: Disabled{}}
		return
	}

	if optUsr.IsNone() {
		slog.Error("user is gone", "id", v.user.ID, "err", err)
		v.user = User{ID: v.user.ID, Status: Disabled{}}
		return
	}

	v.user = optUsr.Unwrap()

	if v.user.Contact.DisplayLanguage == "und" {
		v.locale = language.English
	} else {
		tag, err := language.Parse(v.user.Contact.DisplayLanguage)
		if err != nil {
			// this is just way to verbose and common not to have such value
			// let us simply ignore it.
			//slog.Error("cannot parse user preferred language", "id", v.user.ID, "err", err)
		}

		v.locale = tag
	}
}

func (v *viewImpl) HasResourcePermission(name rebac.Namespace, id rebac.Instance, p permission.ID) bool {
	if !v.Valid() {
		return false
	}

	if v.HasPermission(p) {
		return true
	}

	ok, err := v.rdb.Contains(rebac.Triple{
		Source: rebac.Entity{
			Namespace: Namespace,
			Instance:  rebac.Instance(v.ID()),
		},
		Relation: rebac.Relation(p),
		Target: rebac.Entity{
			Namespace: name,
			Instance:  id,
		},
	})

	if err != nil {
		slog.Error("cannot check resource permission", "err", err)
		return false
	}

	return ok
}

func (v *viewImpl) AuditResource(name rebac.Namespace, id rebac.Instance, p permission.ID) error {
	if !v.HasResourcePermission(name, id, p) {
		var permName = string(p)
		if perm, ok := permission.Find(p); ok {
			permName = perm.Name
		}

		return PermissionDeniedError(permName)
	}

	return nil
}

func (v *viewImpl) Audit(perm permission.ID) error {
	usr := v.refresh()

	if v.user.ID == "" {
		return InvalidSubjectErr
	}

	if !v.Valid() {
		if !usr.EMailVerified {
			return std.NewLocalizedError("Keine Berechtigung", "Die Mail-Adresse zum Konto muss zuerst bestätigt werden.")
		}

		return std.NewLocalizedError("Keine Berechtigung", "Das Nutzerkonto ist nicht gültig.")
	}

	if !v.HasPermission(perm) {
		var name = string(perm)
		if p, ok := permission.Find(perm); ok {
			name = p.Name
		}

		return PermissionDeniedError(name)
	}

	return nil
}

func (v *viewImpl) HasPermission(permission permission.ID) bool {
	v.refresh()

	if !v.Valid() {
		return false
	}

	ok, err := v.rdb.Contains(rebac.Triple{
		Source: rebac.Entity{
			Namespace: Namespace,
			Instance:  rebac.Instance(v.ID()),
		},
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

func (v *viewImpl) ID() ID {
	// security note: it is important, that this implementation always returns
	// a non-empty user id, otherwise there may be implementations which will start with a
	// check that its valid and use this getter for future comparison which will start breaking.
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.user.ID
}

func (v *viewImpl) Avatar() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return string(v.user.Contact.Avatar)
}

func (v *viewImpl) Name() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.user.Contact.Firstname == "" || v.user.Contact.Lastname == "" {
		if v.user.Contact.Firstname != "" {
			return v.user.Contact.Firstname
		}

		return v.user.Contact.Lastname
	}

	return v.user.Contact.Firstname + " " + v.user.Contact.Lastname
}

func (v *viewImpl) Firstname() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.user.Contact.Firstname
}

func (v *viewImpl) Lastname() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.user.Contact.Lastname
}

func (v *viewImpl) Email() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return string(v.user.Email)
}

func (v *viewImpl) Roles() iter.Seq[role.ID] {
	v.refresh()

	if !v.Valid() {
		return func(yield func(role.ID) bool) {}
	}

	return func(yield func(role.ID) bool) {
		it := v.rdb.Query(
			rebac.Select().
				Where().Source().IsNamespace(role.Namespace).
				Where().Relation().Has(rebac.Member).
				Where().Target().Is(Namespace, rebac.Instance(v.ID())),
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

func (v *viewImpl) HasRole(id role.ID) bool {
	v.refresh()

	if !v.Valid() {
		return false
	}

	ok, err := v.rdb.Contains(rebac.Triple{
		Source: rebac.Entity{
			Namespace: role.Namespace,
			Instance:  rebac.Instance(id),
		},
		Relation: rebac.Member,
		Target: rebac.Entity{
			Namespace: Namespace,
			Instance:  rebac.Instance(v.ID()),
		},
	})

	if err != nil {
		slog.Error("cannot check resource permission", "err", err)
		return false
	}

	return ok
}

func (v *viewImpl) Groups() iter.Seq[group.ID] {
	v.refresh()

	if !v.Valid() {
		return func(yield func(group.ID) bool) {}
	}

	return func(yield func(group.ID) bool) {
		it := v.rdb.Query(
			rebac.Select().
				Where().Source().IsNamespace(group.Namespace).
				Where().Relation().Has(rebac.Member).
				Where().Target().Is(Namespace, rebac.Instance(v.ID())),
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

func (v *viewImpl) HasGroup(id group.ID) bool {
	v.refresh()

	if !v.Valid() {
		return false
	}

	ok, err := v.rdb.Contains(rebac.Triple{
		Source: rebac.Entity{
			Namespace: group.Namespace,
			Instance:  rebac.Instance(id),
		},
		Relation: rebac.Member,
		Target: rebac.Entity{
			Namespace: Namespace,
			Instance:  rebac.Instance(v.ID()),
		},
	})

	if err != nil {
		slog.Error("cannot check resource permission", "err", err)
		return false
	}

	return ok
}

func (v *viewImpl) Valid() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.user.EMailVerified && v.user.Enabled()
}

func (v *viewImpl) Language() language.Tag {
	v.refresh()

	return v.locale
}

func (v *viewImpl) SetLanguage(tag language.Tag) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.locale = tag
}

func (v *viewImpl) Bundle() *i18n.Bundle {
	return v.bundle.Load()
}

func (v *viewImpl) SetBundle(b *i18n.Bundle) {
	v.bundle.Store(b)
}

// WithContext wraps the subject, delegates all calls but returns the given context instead.
func WithContext(subject Subject, ctx context.Context) Subject {
	return ctxSubject{ctx: ctx, Subject: subject}
}

type ctxSubject struct {
	ctx context.Context
	Subject
}

func (s ctxSubject) Context() context.Context {
	return s.ctx
}
