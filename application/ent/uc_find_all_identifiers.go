package ent

import (
	"iter"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewFindAllIdentifiers[T Aggregate[T, ID], ID ~string](opts Options, perms Permissions, repo data.Repository[T, ID]) FindAllIdentifiers[T, ID] {
	return func(subject auth.Subject) iter.Seq2[ID, error] {
		return func(yield func(ID, error) bool) {
			for id, err := range repo.Identifiers() {
				if err != nil {
					if !yield("", err) {
						return
					}

					continue
				}

				if !subject.HasResourcePermission(repo.Name(), string(id), perms.FindByID) {
					continue
				}

				if !yield(id, nil) {
					return
				}
			}
		}
	}
}
