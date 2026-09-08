// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"iter"
	"sync"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"golang.org/x/text/language"
)

var (
	StrResName = i18n.MustString("nago.role.resources.name", i18n.Values{language.German: "Rollen", language.English: "Roles"})
	StrResDesc = i18n.MustString("nago.role.resources.desc", i18n.Values{language.German: "Rollen und assoziierte Berechtigungen.", language.English: "Roles and associated permissions."})

	StrLabelSystem = i18n.MustString("nago.role.label.system", i18n.Values{language.German: "Systemrolle", language.English: "System role"})
)

const Namespace rebac.Namespace = "nago.iam.role"

type ID string

type Role struct {
	ID          ID     `json:"id,omitempty" label:"nago.common.label.identifier"`
	Name        string `json:"name,omitempty" label:"nago.common.label.name"`
	Description string `json:"description,omitempty" label:"nago.common.label.description" lines:"3"`

	// System marks a role that is owned and maintained by the application itself, declared through
	// [Configurator.DeclareSystemRole]. Such a role exists so that an operator only has to assign it: its
	// permission set is part of the feature it belongs to, not a configuration decision.
	//
	// It may therefore neither be deleted nor have its permissions replaced (see [Delete] and
	// [UpdatePermissions]); assigning members is unaffected, since that is the entire point of the role.
	// The flag itself is owned by the declaration and cannot be set or cleared through [Update].
	System bool `json:"system,omitempty" label:"nago.role.label.system"`
}

// IsSystem reports whether this role is maintained by the application and thus protected against deletion and
// permission edits.
func (r Role) IsSystem() bool {
	return r.System
}

func (r Role) String() string {
	return r.Name
}

func (r Role) WithIdentity(id ID) Role {
	r.ID = id
	return r
}

func (r Role) Identity() ID {
	return r.ID
}

type Repository = data.Repository[Role, ID]

type FindByID func(subject permission.Auditable, id ID) (option.Opt[Role], error)
type FindAll func(subject permission.Auditable) iter.Seq2[Role, error]
type Create func(subject permission.Auditable, role Role) (ID, error)

// Upsert inserts or updates the given Role by ID. It only returns an error, if the permissions [PermUpdate] and
// [PermCreate] are not available or any other infrastructure error occurs.
type Upsert func(subject permission.Auditable, role Role) (ID, error)
type Update func(subject permission.Auditable, role Role) error
type Delete func(subject permission.Auditable, id ID) error

// FindMyRoles returns only those roles, in which the subject is a member.
type FindMyRoles func(subject permission.Auditable) iter.Seq2[Role, error]

// ListPermissions returns all permissions assigned to the given role.
type ListPermissions func(subject permission.Auditable, id ID) iter.Seq2[permission.ID, error]

// UpdatePermissions replaces all permissions assigned to the given role.
type UpdatePermissions func(subject permission.Auditable, id ID, permissions []permission.ID) error

// UpsertPermissions ensures that at least the given permissions are assigned to the role. Permissions the
// role already holds beyond the given ones are kept; nothing is ever revoked. When every given permission is
// already granted, no write is performed at all, which makes it cheap to call on every application start.
//
// This is the use case behind [Configurator.DeclareSystemRole]: a module declares the permissions its feature
// needs and stays indifferent to whatever else an operator has put into the role.
type UpsertPermissions func(subject permission.Auditable, id ID, permissions []permission.ID) error

type UseCases struct {
	FindByID          FindByID
	FindAll           FindAll
	Create            Create
	Upsert            Upsert
	Update            Update
	Delete            Delete
	FindMyRoles       FindMyRoles
	ListPermissions   ListPermissions
	UpdatePermissions UpdatePermissions
	UpsertPermissions UpsertPermissions
	Resources         rebac.Resources
}

func NewUseCases(repo Repository, bus events.Bus, rdb *rebac.DB) UseCases {
	// note, that we cannot refactor to use auth.Decorate the repo, due to bootstrapping and cycle problem
	var roleMutex sync.Mutex
	findByIdFn := NewFindByID(repo)
	findAllFn := NewFindAll(repo)
	createFn := NewCreate(&roleMutex, repo, bus)
	upsertFn := NewUpsert(&roleMutex, repo, bus)
	updateFn := NewUpdate(&roleMutex, repo, bus)
	deleteFn := NewDelete(&roleMutex, repo, bus, rdb)
	findMyRolesFn := NewFindMyRoles(repo)

	return UseCases{
		FindByID:          findByIdFn,
		FindAll:           findAllFn,
		Create:            createFn,
		Upsert:            upsertFn,
		Update:            updateFn,
		Delete:            deleteFn,
		FindMyRoles:       findMyRolesFn,
		ListPermissions:   NewListPermissions(rdb),
		UpdatePermissions: NewUpdatePermissions(repo, rdb),
		UpsertPermissions: NewUpsertPermissions(rdb),
		Resources:         rebac.NewRepositoryResources(StrResName, StrResDesc, repo),
	}
}
