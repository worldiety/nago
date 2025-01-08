package user

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewUpdateMyContact(mutex *sync.Mutex, repo Repository) UpdateMyContact {
	return func(subject AuditableUser, contact Contact) error {
		if !subject.Valid() {
			// bootstrap error message
			return std.NewLocalizedError("Nicht eingeloggt", "Diese Funktion steht nur eingeloggten Nutzern zur Verfügung.")
		}

		// mutex is important, otherwise we may re-create a user accidentally
		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := repo.FindByID(subject.ID())
		if err != nil {
			return fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return std.NewLocalizedError("Nicht eingeloggt", "Der Nutzer ist nicht (mehr) vorhanden.")
		}

		usr := optUsr.Unwrap()
		usr.Contact = contact
		return repo.Save(usr)
	}
}
