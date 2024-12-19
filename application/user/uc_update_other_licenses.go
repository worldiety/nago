package user

import (
	"fmt"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/pkg/std"
	"slices"
	"sync"
)

func NewUpdateOtherLicenses(mutex *sync.Mutex, repo Repository) UpdateOtherLicenses {
	return func(subject AuditableUser, id ID, licenses []license.ID) error {
		if err := subject.Audit(PermUpdateOtherLicenses); err != nil {
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

		slices.Sort(licenses)
		slices.Compact(licenses)

		usr := optUsr.Unwrap()
		usr.Licenses = licenses
		return repo.Save(usr)
	}
}
