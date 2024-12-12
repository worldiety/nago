package auth

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"strings"
	"sync"
)

type Aggregate[A any, ID comparable] interface {
	data.Aggregate[ID]
	WithIdentity(ID) A
}

type DecoratedRepository[E Aggregate[E, EID], EID data.IDType] struct {
	repo           data.Repository[E, EID]
	PermFindByID   permission.ID
	PermFindAll    permission.ID
	PermDeleteByID permission.ID
	PermCreate     permission.ID
	PermUpdate     permission.ID

	FindByID func(subject Subject, id EID) (std.Option[E], error)
	FindAll  func(subject Subject) iter.Seq2[E, error]

	// DeleteByID removes the associated entity. It does also succeed if the identifier does not exist.
	DeleteByID func(subject Subject, id EID) error

	// Create saves the given entity under the assumption that it does not exist. If it does exist,
	// an error is returned. A zero Identity is considered empty and a new ID is generated and returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Create func(subject Subject, e E) (EID, error)

	// Update saves the given entity under the assumption that it already exists. If it does not exist,
	// an error is returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Update func(subject Subject, e E) error

	// Upsert either updates or creates a new entity. It is not an error, if an entity already exists or not.
	// A zero value Identity is considered as Create request and a new ID is generated and returned.
	// This is only guaranteed to work within a single process and database. Keep in mind, that shared
	// databases may cause ghost updates and other kinds of logical races due to eventual consistency.
	Upsert func(subject Subject, e E) (EID, error)
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

func DecorateRepository[E Aggregate[E, EID], EID ~string](opts DecoratorOptions, repo data.Repository[E, EID]) *DecoratedRepository[E, EID] {
	var repoMutex sync.Mutex

	if opts.PermissionPrefix == "" {
		panic("permission prefix not set")
	}

	if strings.HasSuffix(string(opts.PermissionPrefix), ".") {
		opts.PermissionPrefix += "."
	}

	crud := &DecoratedRepository[E, EID]{
		repo:           repo,
		PermFindByID:   permission.Declare[_FindByID[E]](opts.PermissionPrefix+"find_by_id", opts.EntityName+" finden", "Träger dieser Berechtigung können Elemente vom Typ '"+opts.EntityName+"' über die eindeutige ID anzeigen."),
		PermFindAll:    permission.Declare[_FindAll[E]](opts.PermissionPrefix+"find_all", "Alle vom Typ '"+opts.EntityName+"' finden", "Träger dieser Berechtigung können alle Elemente vom Typ '"+opts.EntityName+"' entdecken und anzeigen."),
		PermCreate:     permission.Declare[_Create[E]](opts.PermissionPrefix+"create", opts.EntityName+" erstellen", "Träger dieser Berechtigung können neue Elemente vom Typ '"+opts.EntityName+"' anlegen aber keine vorhandenen aktualisieren."),
		PermUpdate:     permission.Declare[_Update[E]](opts.PermissionPrefix+"update", opts.EntityName+" aktualisieren", "Träger dieser Berechtigung können vorhandene Elemente vom Typ '"+opts.EntityName+"' aktualisieren aber keine neu anlegen."),
		PermDeleteByID: permission.Declare[_DeleteByID[E]](opts.PermissionPrefix+"delete_by_id", opts.EntityName+" löschen", "Träger dieser Berechtigung können Elemente vom Typ '"+opts.EntityName+"' über die eindeutige ID löschen."),
	}

	crud.FindByID = func(subject Subject, id EID) (std.Option[E], error) {
		if err := subject.Audit(crud.PermFindByID); err != nil {
			return std.None[E](), err
		}

		return crud.repo.FindByID(id)
	}

	crud.FindAll = func(subject Subject) iter.Seq2[E, error] {
		if err := subject.Audit(crud.PermFindAll); err != nil {
			return xiter.WithError[E](err)
		}

		return crud.repo.All()
	}

	crud.DeleteByID = func(subject Subject, id EID) error {
		if err := subject.Audit(crud.PermDeleteByID); err != nil {
			return err
		}

		repoMutex.Lock() // this only works for in-process
		defer repoMutex.Unlock()

		return repo.DeleteByID(id)
	}

	crud.Upsert = func(subject Subject, e E) (EID, error) {
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

	crud.Update = func(subject Subject, e E) error {
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

	crud.Create = func(subject Subject, e E) (EID, error) {
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
