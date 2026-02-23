package ent

import (
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewDeleteByID[T Aggregate[T, ID], ID ~string](opts Options, perms Permissions, repo data.Repository[T, ID]) DeleteByID[T, ID] {
	return func(subject auth.Subject, id ID) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(id), perms.DeleteByID); err != nil {
			return err
		}

		return repo.DeleteByID(id)
	}
}
