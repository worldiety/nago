// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package group

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"sync"
)

// ID of a Group.
type ID string

type Group struct {
	ID          ID     `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (r Group) WithIdentity(id ID) Group {
	r.ID = id
	return r
}

func (r Group) Identity() ID {
	return r.ID
}

func (r Group) String() string {
	if r.Name != "" {
		return r.Name
	}

	return string(r.ID)
}

type Repository = data.Repository[Group, ID]

// note, that we are in a bootstrapping condition here and cannot refer to auth/user.Subject due to dep cycles, therefore we depend on the Auditable contract

type FindByID func(subject permission.Auditable, id ID) (std.Option[Group], error)
type FindAll func(subject permission.Auditable) iter.Seq2[Group, error]
type Create func(subject permission.Auditable, group Group) (ID, error)

// Upsert inserts or updates the given Group by ID. It only returns an error, if the permissions [PermUpdate] and
// [PermCreate] are not available or any other infrastructure error occurs.
type Upsert func(subject permission.Auditable, group Group) (ID, error)
type Update func(subject permission.Auditable, group Group) error
type Delete func(subject permission.Auditable, id ID) error

// FindMyGroups returns only those groups, in which the subject is a member.
type FindMyGroups func(subject permission.Auditable) iter.Seq2[Group, error]

type UseCases struct {
	FindByID     FindByID
	FindAll      FindAll
	Create       Create
	Upsert       Upsert
	Update       Update
	Delete       Delete
	FindMyGroups FindMyGroups
}

func NewUseCases(repo Repository) UseCases {
	var mutex sync.Mutex

	findByIdFn := NewFindByID(repo)
	findAllFn := NewFindAll(repo)
	createFn := NewCreate(&mutex, repo)
	upsertFn := NewUpsert(&mutex, repo)
	updateFn := NewUpdate(&mutex, repo)
	deleteFn := NewDelete(&mutex, repo)
	findMyGroupsFn := NewFindMyGroups(repo)

	return UseCases{
		FindByID:     findByIdFn,
		FindAll:      findAllFn,
		Create:       createFn,
		Upsert:       upsertFn,
		Update:       updateFn,
		Delete:       deleteFn,
		FindMyGroups: findMyGroupsFn,
	}
}
