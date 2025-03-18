package template

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
)

func NewFindByID(repo Repository) FindByID {
	return func(subject auth.Subject, id ID) (option.Opt[Project], error) {
		if err := subject.AuditResource(repo.Name(), string(id), PermFindByID); err != nil {
			return option.Opt[Project]{}, err
		}

		return repo.FindByID(id)
	}
}
