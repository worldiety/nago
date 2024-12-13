package group

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindByID(repo Repository) FindByID {
	return func(subject permission.Auditable, id ID) (std.Option[Group], error) {
		if err := subject.Audit(PermFindByID); err != nil {
			return std.None[Group](), err
		}

		return repo.FindByID(id)
	}
}
