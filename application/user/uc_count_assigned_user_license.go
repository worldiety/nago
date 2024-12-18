package user

import (
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"sync"
)

func NewCountAssignedUserLicense(mutex *sync.Mutex, users Repository) CountAssignedUserLicense {
	return func(auditable permission.Auditable, id license.ID) (int, error) {
		if err := auditable.Audit(PermCountAssignedUserLicense); err != nil {
			return 0, err
		}

		// mutex is not that important when reading, however, let us count at least a consistent point-in-time
		mutex.Lock()
		defer mutex.Unlock()

		var count int
		for user, err := range users.All() {
			if err != nil {
				return 0, err
			}

			for _, lic := range user.Licenses {
				if lic == id {
					count++
				}
			}
		}

		return count, nil
	}
}
