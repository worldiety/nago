package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"sync"
)

func NewUpsertUserLicense(mutex *sync.Mutex, repo UserLicenseRepository) UpsertUserLicense {
	return func(subject permission.Auditable, license UserLicense) (ID, error) {
		if err := subject.Audit(PermCreateUserLicense); err != nil {
			return "", err
		}

		if err := subject.Audit(PermUpdateUserLicense); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if license.ID == "" {
			license.ID = data.RandIdent[ID]()
		}

		return license.ID, repo.Save(license)
	}
}
