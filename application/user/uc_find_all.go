package user

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindAll(repository Repository) FindAll {
	return func(subject permission.Auditable) iter.Seq2[User, error] {
		if err := subject.Audit(PermFindAll); err != nil {
			return xiter.WithError[User](err)
		}

		return repository.All()
	}
}
