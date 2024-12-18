package license

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewUpdateUserLicense(mutex *sync.Mutex, repo UserLicenseRepository) UpdateUserLicense {
	return func(subject permission.Auditable, license UserLicense) error {
		if err := subject.Audit(PermUpdateUserLicense); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optE, err := repo.FindByID(license.ID)
		if err != nil {
			return err
		}

		if optE.IsNone() {
			return std.NewLocalizedError("User-Lizenz nicht aktualisierbar", fmt.Sprintf("Die Lizenz mit der ID '%v' existiert nicht.", license.ID))
		}

		return repo.Save(license)
	}
}
