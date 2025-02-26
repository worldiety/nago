package user

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindAllIdentifiers(repository Repository) FindAllIdentifiers {
	return func(subject permission.Auditable) iter.Seq2[ID, error] {
		if err := subject.Audit(PermFindAll); err != nil {
			return xiter.WithError[ID](err)
		}

		return repository.Identifiers()
	}
}
