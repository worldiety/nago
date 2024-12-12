package group

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

type Repository = data.Repository[Group, ID]

type FindByID func(subject permission.Auditable, id ID) (std.Option[Group], error)
type FindAll func(subject permission.Auditable) iter.Seq2[Group, error]
type Create func(subject permission.Auditable, group Group) error

// Upsert inserts or updates the given Group by ID. It only returns an error, if the permissions [PermUpdate] and
// [PermCreate] are not available or any other infrastructure error occurs.
type Upsert func(subject permission.Auditable, group Group) error
type Update func(subject permission.Auditable, group Group) error
type Delete func(subject permission.Auditable, id ID) error

func NewFindByID(repo Repository) FindByID {
	return func(subject permission.Auditable, id ID) (std.Option[Group], error) {
		if err := subject.Audit(PermFindByID); err != nil {
			return std.None[Group](), err
		}

		return repo.FindByID(id)
	}
}

func NewFindAll(repo Repository) FindAll {
	return func(subject permission.Auditable) iter.Seq2[Group, error] {
		if err := subject.Audit(PermFindAll); err != nil {
			return xiter.WithError[Group](err)
		}

		return repo.All()
	}
}

func NewCreate(mutex *sync.Mutex, repo Repository) Create {
	return func(subject permission.Auditable, group Group) error {
		if err := subject.Audit(PermCreate); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if strings.TrimSpace(string(group.ID)) == "" {
			group.ID = data.RandIdent[ID]()
		}

		optGroup, err := repo.FindByID(group.ID)
		if err != nil {
			return fmt.Errorf("cannot find group by id: %w", err)
		}

		if optGroup.IsSome() {
			return std.NewLocalizedError("Ungültige EID", "Eine Gruppe mit derselben EID ist bereits vorhanden.")
		}

		return repo.Save(group)
	}
}

func NewUpsert(mutex *sync.Mutex, repo Repository) Upsert {
	return func(subject permission.Auditable, group Group) error {
		if err := subject.Audit(PermCreate); err != nil {
			return err
		}

		if err := subject.Audit(PermUpdate); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		createNew := false
		if strings.TrimSpace(string(group.ID)) == "" {
			group.ID = data.RandIdent[ID]()
			createNew = true
		}

		optGroup, err := repo.FindByID(group.ID)
		if err != nil {
			return fmt.Errorf("cannot find group by id: %w", err)
		}

		if optGroup.IsSome() && createNew {
			return fmt.Errorf("random id collision on upsert creation")
		}

		return repo.Save(group)
	}
}

func NewUpdate(mutex *sync.Mutex, repo Repository) Update {
	return func(subject permission.Auditable, group Group) error {
		if err := subject.Audit(PermUpdate); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if strings.TrimSpace(string(group.ID)) == "" {
			return std.NewLocalizedError("Ungültige EID", "Eine leere Gruppen EID ist nicht zulässig.")
		}

		return repo.Save(group)
	}
}

func NewDelete(mutex *sync.Mutex, repo Repository) Delete {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDelete); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteByID(id)
	}
}
