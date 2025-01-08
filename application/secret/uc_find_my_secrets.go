package secret

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"slices"
)

func NewFindMySecrets(repository Repository) FindMySecrets {
	return func(subject auth.Subject) iter.Seq2[Secret, error] {
		if err := subject.Audit(PermFindMySecrets); err != nil {
			return xiter.WithError[Secret](err)
		}

		return func(yield func(Secret, error) bool) {
			for secret, err := range repository.All() {
				if err != nil {
					if !yield(Secret{}, err) {
						return
					}
					continue
				}

				if slices.Contains(secret.Owners, subject.ID()) {
					if !yield(secret, nil) {
						return
					}
				}
			}
		}
	}
}
