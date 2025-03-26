// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package rcrud provides simple but automatic use case implementation for create, read, update and delete
// based on a given repository. The required permissions are also automatically registered.
package rcrud

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"reflect"
	"strings"
	"sync"
)

type Aggregate[A any, ID comparable] interface {
	data.Aggregate[ID]
	WithIdentity(ID) A
}

// UseCases represent the most basic and simplest CRUD-based use cases. See also [NewUseCases] to automatically
// derive an instance from a [data.Repository]. This is probably only useful for rapid prototyping for
// the most simple CRUD use cases.
type UseCases[E Aggregate[E, ID], ID ~string] interface {
	PermFindByID() permission.ID
	PermFindAll() permission.ID
	PermDeleteByID() permission.ID
	PermCreate() permission.ID
	PermUpdate() permission.ID

	FindByID(subject auth.Subject, id ID) (std.Option[E], error)
	FindAll(subject auth.Subject) iter.Seq2[E, error]

	// DeleteByID removes the associated entity. It does also succeed if the identifier does not exist.
	DeleteByID(subject auth.Subject, id ID) error

	// Upsert either updates or creates a new entity. It is not an error, if an entity already exists or not.
	// A zero value Identity is considered as Create request and a new ID is generated and returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Upsert(subject auth.Subject, entity E) (ID, error)

	// Create saves the given entity under the assumption that it does not exist. If it does exist,
	// an error is returned. A zero Identity is considered empty and a new ID is generated and returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Create(subject auth.Subject, entity E) (ID, error)

	// Update saves the given entity under the assumption that it already exists. If it does not exist,
	// an error is returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Update(subject auth.Subject, entity E) error
}

type useCasesImpl[E Aggregate[E, EID], EID data.IDType] struct {
	decorated *Funcs[E, EID]
	mutex     *sync.Mutex
}

func (u useCasesImpl[E, EID]) PermFindByID() permission.ID {
	return u.decorated.PermFindByID
}

func (u useCasesImpl[E, EID]) PermFindAll() permission.ID {
	return u.decorated.PermFindAll
}

func (u useCasesImpl[E, EID]) PermDeleteByID() permission.ID {
	return u.decorated.PermDeleteByID
}

func (u useCasesImpl[E, EID]) PermCreate() permission.ID {
	return u.decorated.PermCreate
}

func (u useCasesImpl[E, EID]) PermUpdate() permission.ID {
	return u.decorated.PermUpdate
}

func (u useCasesImpl[E, EID]) FindByID(subject auth.Subject, id EID) (std.Option[E], error) {
	return u.decorated.FindByID(subject, id)
}

func (u useCasesImpl[E, EID]) FindAll(subject auth.Subject) iter.Seq2[E, error] {
	return u.decorated.FindAll(subject)
}

func (u useCasesImpl[E, EID]) DeleteByID(subject auth.Subject, id EID) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	return u.decorated.DeleteByID(subject, id)
}

func (u useCasesImpl[E, EID]) Upsert(subject auth.Subject, entity E) (EID, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.decorated.Upsert != nil {
		return u.decorated.Upsert(subject, entity)
	}

	optE, err := u.FindByID(subject, entity.Identity())
	if err != nil {
		return entity.Identity(), err
	}

	if optE.IsSome() {
		return entity.Identity(), u.decorated.Update(subject, entity)
	}

	return u.decorated.Create(subject, entity)
}

func (u useCasesImpl[E, EID]) Create(subject auth.Subject, entity E) (EID, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.decorated.Create != nil {
		return u.decorated.Create(subject, entity)

	}

	return u.decorated.Upsert(subject, entity)
}

func (u useCasesImpl[E, EID]) Update(subject auth.Subject, entity E) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.decorated.Update != nil {
		return u.decorated.Update(subject, entity)
	}

	_, err := u.decorated.Upsert(subject, entity)
	return err
}

type Funcs[E Aggregate[E, EID], EID data.IDType] struct {
	repo           data.Repository[E, EID]
	PermFindByID   permission.ID
	PermFindAll    permission.ID
	PermDeleteByID permission.ID
	PermCreate     permission.ID
	PermUpdate     permission.ID

	FindByID func(subject auth.Subject, id EID) (std.Option[E], error)
	FindAll  func(subject auth.Subject) iter.Seq2[E, error]

	// DeleteByID removes the associated entity. It does also succeed if the identifier does not exist.
	DeleteByID func(subject auth.Subject, id EID) error

	// Create saves the given entity under the assumption that it does not exist. If it does exist,
	// an error is returned. A zero Identity is considered empty and a new ID is generated and returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Create func(subject auth.Subject, e E) (EID, error)

	// Update saves the given entity under the assumption that it already exists. If it does not exist,
	// an error is returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Update func(subject auth.Subject, e E) error

	// Upsert either updates or creates a new entity. It is not an error, if an entity already exists or not.
	// A zero value Identity is considered as Create request and a new ID is generated and returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Upsert func(subject auth.Subject, e E) (EID, error)
}

type _FindByID[E any] func()
type _FindAll[E any] func()
type _DeleteByID[E any] func()
type _Create[E any] func()
type _Update[E any] func()

type DecoratorOptions struct {
	PermissionPrefix permission.ID
	EntityName       string // translated entity name for auto generated permission details.
}

func UseCasesFrom[E Aggregate[E, EID], EID ~string](funcs *Funcs[E, EID]) UseCases[E, EID] {
	return useCasesImpl[E, EID]{decorated: funcs, mutex: new(sync.Mutex)}
}

func DecorateRepository[E Aggregate[E, EID], EID ~string](opts DecoratorOptions, repo data.Repository[E, EID]) *Funcs[E, EID] {
	var repoMutex sync.Mutex

	if opts.PermissionPrefix == "" {
		panic("permission prefix not set")
	}

	if !opts.PermissionPrefix.Valid() {
		panic(fmt.Errorf("permission prefix is invalid: %v", opts.PermissionPrefix))
	}

	if !strings.HasSuffix(string(opts.PermissionPrefix), ".") {
		opts.PermissionPrefix += "."
	}

	if opts.EntityName == "" {
		opts.EntityName = reflect.TypeFor[E]().Name()
	}

	crud := &Funcs[E, EID]{
		repo:           repo,
		PermFindByID:   permission.Declare[_FindByID[E]](opts.PermissionPrefix+"find_by_id", opts.EntityName+" finden", "Träger dieser Berechtigung können Elemente vom Typ '"+opts.EntityName+"' über die eindeutige ID anzeigen."),
		PermFindAll:    permission.Declare[_FindAll[E]](opts.PermissionPrefix+"find_all", "Alle vom Typ '"+opts.EntityName+"' finden", "Träger dieser Berechtigung können alle Elemente vom Typ '"+opts.EntityName+"' entdecken und anzeigen."),
		PermCreate:     permission.Declare[_Create[E]](opts.PermissionPrefix+"create", opts.EntityName+" erstellen", "Träger dieser Berechtigung können neue Elemente vom Typ '"+opts.EntityName+"' anlegen aber keine vorhandenen aktualisieren."),
		PermUpdate:     permission.Declare[_Update[E]](opts.PermissionPrefix+"update", opts.EntityName+" aktualisieren", "Träger dieser Berechtigung können vorhandene Elemente vom Typ '"+opts.EntityName+"' aktualisieren aber keine neu anlegen."),
		PermDeleteByID: permission.Declare[_DeleteByID[E]](opts.PermissionPrefix+"delete_by_id", opts.EntityName+" löschen", "Träger dieser Berechtigung können Elemente vom Typ '"+opts.EntityName+"' über die eindeutige ID löschen."),
	}

	crud.FindByID = func(subject auth.Subject, id EID) (std.Option[E], error) {
		if err := subject.Audit(crud.PermFindByID); err != nil {
			return std.None[E](), err
		}

		return crud.repo.FindByID(id)
	}

	crud.FindAll = func(subject auth.Subject) iter.Seq2[E, error] {
		if err := subject.Audit(crud.PermFindAll); err != nil {
			return xiter.WithError[E](err)
		}

		return crud.repo.All()
	}

	crud.DeleteByID = func(subject auth.Subject, id EID) error {
		if err := subject.Audit(crud.PermDeleteByID); err != nil {
			return err
		}

		repoMutex.Lock() // this only works for in-process
		defer repoMutex.Unlock()

		return repo.DeleteByID(id)
	}

	crud.Upsert = func(subject auth.Subject, e E) (EID, error) {
		repoMutex.Lock() // this only works for in-process
		defer repoMutex.Unlock()

		var zeroID EID
		if err := subject.Audit(crud.PermUpdate); err != nil {
			return zeroID, err
		}

		if err := subject.Audit(crud.PermCreate); err != nil {
			return zeroID, err
		}

		if e.Identity() == zeroID {
			e = e.WithIdentity(data.RandIdent[EID]())

			// we must check, if the unlikely case of a collision occurred
			optE, err := repo.FindByID(e.Identity())
			if err != nil {
				return zeroID, err
			}

			if optE.IsSome() {
				return zeroID, fmt.Errorf("id collision on %v", e.Identity())
			}
		}

		return e.Identity(), repo.Save(e)
	}

	crud.Update = func(subject auth.Subject, e E) error {
		if err := subject.Audit(crud.PermUpdate); err != nil {
			return err
		}

		repoMutex.Lock() // this only works for in-process
		defer repoMutex.Unlock()

		var zero EID
		if e.Identity() == zero {
			return std.NewLocalizedError("Ungültige ID", "Die Entität hat eine leere ID und kann daher nicht aktualisiert werden.")
		}

		optE, err := repo.FindByID(e.Identity())
		if err != nil {
			return err
		}

		if optE.IsNone() {
			return std.NewLocalizedError("Ungültige ID", "Die Entität mit der ID'"+string(e.Identity())+"' kann nicht aktualisiert werden, da sie nicht (mehr) existiert.")
		}

		return repo.Save(e)
	}

	crud.Create = func(subject auth.Subject, e E) (EID, error) {
		if err := subject.Audit(crud.PermCreate); err != nil {
			return e.Identity(), err
		}

		repoMutex.Lock() // this only works for in-process
		defer repoMutex.Unlock()

		var zeroID EID
		if e.Identity() == zeroID {
			e = e.WithIdentity(data.RandIdent[EID]())
		}

		optE, err := repo.FindByID(e.Identity())
		if err != nil {
			return zeroID, err
		}

		if optE.IsSome() {
			return zeroID, std.NewLocalizedError("Ungültige ID", "Die Entität mit der ID'"+string(e.Identity())+"' kann nicht angelegt werden, da eine andere mit derselben ID bereits existiert.")
		}

		return e.Identity(), repo.Save(e)
	}

	return crud
}
