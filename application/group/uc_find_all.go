package group

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindAll(repo Repository) FindAll {
	return func(subject permission.Auditable) iter.Seq2[Group, error] {
		if err := subject.Audit(PermFindAll); err != nil {
			return xiter.WithError[Group](err)
		}

		return repo.All()
	}
}
