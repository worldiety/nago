package ent

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewFindByID[T Aggregate[T, ID], ID ~string](opts Options, perms Permissions, repo data.Repository[T, ID]) FindByID[T, ID] {
	return func(subject auth.Subject, id ID) (option.Opt[T], error) {
		if err := subject.AuditResource(repo.Name(), string(id), perms.FindByID); err != nil {
			return option.Opt[T]{}, err
		}

		return repo.FindByID(id)
	}
}
