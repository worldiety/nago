package user

import (
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"slices"
	"sync"
)

func NewRevokeAssignedUserLicense(mutex *sync.Mutex, users Repository) RevokeAssignedUserLicense {
	return func(auditable permission.Auditable, id license.ID, count int) error {
		if err := auditable.Audit(PermRevokeAssignedUserLicense); err != nil {
			return err
		}

		// mutex is important, otherwise we may revoke an inconsistent amount and/or re-create user when saving
		mutex.Lock()
		defer mutex.Unlock()

		for user, err := range users.All() {
			if err != nil {
				return err
			}

			if count == 0 {
				// we are done
				return nil
			}

			if slices.Contains(user.Licenses, id) {
				if count < 0 {
					// remove-all case
					user.Licenses = slices.DeleteFunc(user.Licenses, func(id license.ID) bool {
						return id == id
					})
				} else {
					// remove-one case
					user.Licenses = slices.DeleteFunc(user.Licenses, func(id license.ID) bool {
						return id == id
					})
					count--
				}

				// mutated user, save it
				if err := users.Save(user); err != nil {
					return err
				}
			}

		}

		return nil
	}
}
