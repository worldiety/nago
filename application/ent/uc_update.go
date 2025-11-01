package ent

import (
	"fmt"
	"os"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewUpdate[T Aggregate[T, ID], ID ~string](opts Options, perms Permissions, repo data.Repository[T, ID]) Update[T, ID] {
	return func(subject auth.Subject, entity T) error {
		if err := subject.AuditResource(repo.Name(), string(entity.Identity()), perms.Update); err != nil {
			return err
		}

		if opts.Mutex != nil {
			opts.Mutex.Lock()
			defer opts.Mutex.Unlock()
		}

		if entity.Identity() == "" {
			return fmt.Errorf("cannot update entity: id is required and must not be empty")
		}

		if optEnt, err := repo.FindByID(entity.Identity()); optEnt.IsNone() || err != nil {
			if err != nil {
				return fmt.Errorf("cannot check repository for existing entity: %w", err)
			}

			return fmt.Errorf("entity to update does exist: %s: %w", entity.Identity(), os.ErrNotExist)
		}

		return repo.Save(entity)
	}
}
