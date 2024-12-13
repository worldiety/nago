package workspace

import (
	"go.wdy.de/nago/auth"
)

var permRemove = annotation.Permission[Remove]("de.worldiety.nago.workspace.remove")

type Remove func(subject auth.Subject, id ID) error

func NewRemove(repo Repository) Remove {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(permRemove.Identity()); err != nil {
			return err
		}

		return repo.DeleteByID(id)
	}
}
