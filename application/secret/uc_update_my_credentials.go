package secret

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"slices"
	"sync"
	"time"
)

func NewUpdateMyCredentials(mutex *sync.Mutex, repository Repository) UpdateMyCredentials {
	return func(subject auth.Subject, id ID, credentials Credentials) error {
		if err := subject.Audit(PermUpdateMyCredentials); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optSecret, err := repository.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find secret by id: %w", err)
		}

		if optSecret.IsNone() {
			return std.NewLocalizedError("Secret Credentials nicht aktualisiert", fmt.Sprintf("Das Secret existiert nicht: %v", id))
		}

		secret := optSecret.Unwrap()
		if !slices.Contains(secret.Owners, subject.ID()) {
			return fmt.Errorf("secret not owned by subject: %v", subject.ID())
		}

		// should we check and update group and membership details here?
		secret.Credentials = credentials
		secret.LastMod = time.Now()

		return repository.Save(secret)
	}
}
