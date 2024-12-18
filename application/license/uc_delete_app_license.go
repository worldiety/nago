package license

import (
	"go.wdy.de/nago/application/permission"
	"sync"
)

func NewDeleteAppLicense(mutex *sync.Mutex, repo AppLicenseRepository) DeleteAppLicense {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDeleteAppLicense); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteByID(id)
	}
}
