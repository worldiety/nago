// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"iter"
	"sync"
)

type ID string

type Role struct {
	ID          ID              `json:"id,omitempty" table-visible:"false"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty" label:"Beschreibung"`
	Permissions []permission.ID `json:"permissions,omitempty" source:"nago.permissions" label:"Berechtigungen" table-visible:"false"`
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

type UseCases struct {
	FindByID    FindByID
	FindAll     FindAll
	Create      Create
	Upsert      Upsert
	Update      Update
	Delete      Delete
	FindMyRoles FindMyRoles
}

func NewUseCases(repo Repository, bus events.Bus) UseCases {
	// note, that we cannot refactor to use auth.Decorate the repo, due to bootstrapping and cycle problem
	var roleMutex sync.Mutex
	findByIdFn := NewFindByID(repo)
	findAllFn := NewFindAll(repo)
	createFn := NewCreate(&roleMutex, repo, bus)
	upsertFn := NewUpsert(&roleMutex, repo, bus)
	updateFn := NewUpdate(&roleMutex, repo, bus)
	deleteFn := NewDelete(&roleMutex, repo, bus)
	findMyRolesFn := NewFindMyRoles(repo)

	return UseCases{
		FindByID:    findByIdFn,
		FindAll:     findAllFn,
		Create:      createFn,
		Upsert:      upsertFn,
		Update:      updateFn,
		Delete:      deleteFn,
		FindMyRoles: findMyRolesFn,
	}
}
