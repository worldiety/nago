package secret

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"slices"
)

func NewFindGroupSecrets(repository Repository) FindGroupSecrets {
	return func(subject auth.Subject, gid group.ID) iter.Seq2[Secret, error] {
		if !subject.HasGroup(gid) {
			return xiter.WithError[Secret](user.InvalidSubjectErr)
		}

		return func(yield func(Secret, error) bool) {
			for secret, err := range repository.All() {
				if err != nil {
					if !yield(Secret{}, err) {
						return
					}
				}

				if slices.Contains(secret.Groups, gid) {
					if !yield(secret, err) {
						return
					}
				}
			}
		}

	}
}
