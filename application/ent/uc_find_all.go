package ent

import (
	"iter"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewFindAll[T Aggregate[T, ID], ID ~string](opts Options, perms Permissions, repo data.Repository[T, ID]) FindAll[T, ID] {
	return func(subject auth.Subject) iter.Seq2[T, error] {
		return func(yield func(T, error) bool) {
			var zero T
			for e, err := range repo.All() {
				if err != nil {
					if !yield(zero, err) {
						return
					}

					continue
				}

				if !subject.HasResourcePermission(rebac.Namespace(repo.Name()), rebac.Instance(e.Identity()), perms.FindAll) {
					continue
				}

				if !yield(e, nil) {
					return
				}
			}
		}
	}
}
