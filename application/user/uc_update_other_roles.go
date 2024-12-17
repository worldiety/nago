package user

import (
	"fmt"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewUpdateOtherRoles(mutex *sync.Mutex, repo Repository) UpdateOtherRoles {
	return func(subject AuditableUser, id ID, roles []role.ID) error {
		if err := subject.Audit(PermUpdateOtherContact); err != nil {
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
		usr.Roles = roles
		return repo.Save(usr)
	}
}
