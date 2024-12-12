package user

import (
	"go.wdy.de/nago/application/permission"
)

func NewDelete(repository Repository) Delete {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDelete); err != nil {
			return err
		}

		return repository.DeleteByID(id)
	}
}
