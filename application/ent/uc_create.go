package ent

import (
	"fmt"
	"os"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewCreate[T Aggregate[T, ID], ID ~string](opts Options, perms Permissions, repo data.Repository[T, ID]) Create[T, ID] {
	return func(subject auth.Subject, entity T) (ID, error) {
		if err := subject.Audit(perms.Create); err != nil {
			return "", err
		}

		if opts.Mutex != nil {
			opts.Mutex.Lock()
			defer opts.Mutex.Unlock()
		}

		if entity.Identity() == "" {
			id := data.RandIdent[ID]()
			entity = entity.WithIdentity(id)
			if entity.Identity() != id {
				panic(fmt.Errorf("implementation failure of %T.WithIdentity(): identity has not been set", entity))
			}
		}

		if optEnt, err := repo.FindByID(entity.Identity()); optEnt.IsSome() || err != nil {
			if err != nil {
				return "", fmt.Errorf("cannot check repository for existing entity: %w", err)
			}

			return "", fmt.Errorf("entity already exists: %s: %w", entity.Identity(), os.ErrExist)
		}

		if err := repo.Save(entity); err != nil {
			return "", fmt.Errorf("cannot save entity: %w", err)
		}

		return entity.Identity(), nil
	}
}
