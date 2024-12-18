package license

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewCreateUserLicense(mutex *sync.Mutex, repo UserLicenseRepository) CreateUserLicense {
	return func(subject permission.Auditable, license UserLicense) (ID, error) {
		if err := subject.Audit(PermCreateUserLicense); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if license.ID == "" {
			license.ID = data.RandIdent[ID]()
		}

		optE, err := repo.FindByID(license.ID)
		if err != nil {
			return "", err
		}

		if optE.IsSome() {
			return "", std.NewLocalizedError("User-Lizenz nicht erstellbar", fmt.Sprintf("Die Lizenz mit der ID '%v' existiert bereits.", license.ID))
		}

		return license.ID, repo.Save(license)
	}
}
