package license

import (
	"go.wdy.de/nago/application/permission"
	"sync"
)

func NewDeleteUserLicense(mutex *sync.Mutex, repo UserLicenseRepository) DeleteUserLicense {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDeleteUserLicense); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteByID(id)
	}
}
