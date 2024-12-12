package user

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindByID(repository Repository) FindByID {
	return func(subject permission.Auditable, id ID) (std.Option[User], error) {
		if err := subject.Audit(PermFindByID); err != nil {
			return std.None[User](), err
		}

		return repository.FindByID(id)
	}
}
