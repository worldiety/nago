package user

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewUpdateOtherGroups(mutex *sync.Mutex, repo Repository) UpdateOtherGroups {
	return func(subject AuditableUser, id ID, groups []group.ID) error {
		if err := subject.Audit(PermUpdateOtherPermissions); err != nil {
			return err
		}

		// mutex is important, otherwise we may re-create a user accidentally
		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return std.NewLocalizedError("Nutzer nicht aktualisiert", "Der Nutzer ist nicht (mehr) vorhanden.")
		}

		usr := optUsr.Unwrap()
		usr.Groups = groups
		return repo.Save(usr)
	}
}
