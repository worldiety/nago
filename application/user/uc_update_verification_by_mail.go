package user

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewUpdateVerificationByMail(mutex *sync.Mutex, repo Repository, byMail FindByMail) UpdateVerificationByMail {
	return func(subject permission.Auditable, mail Email, verified bool) error {
		if err := subject.Audit(PermUpdateOtherContact); err != nil {
			return err
		}

		// mutex is important, otherwise we may re-create a user accidentally
		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := byMail(subject, mail)
		if err != nil {
			return fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return std.NewLocalizedError("Nutzer nicht aktualisiert", "Der Nutzer ist nicht (mehr) vorhanden.")
		}

		usr := optUsr.Unwrap()
		usr.VerificationCode = Code{}
		usr.EMailVerified = true
		return repo.Save(usr)
	}
}
