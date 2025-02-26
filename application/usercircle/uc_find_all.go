package usercircle

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindAll(repo Repository) FindAll {
	return func(subject auth.Subject) iter.Seq2[Circle, error] {
		if err := subject.Audit(PermFindAll); err != nil {
			return xiter.WithError[Circle](err)
		}

		return repo.All()
	}
}
