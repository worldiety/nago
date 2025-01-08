package secret

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"sync"
	"time"
)

func NewCreateSecret(mutex *sync.Mutex, secrets Repository) CreateSecret {
	return func(subject auth.Subject, credentials Credentials) (ID, error) {
		if err := subject.Audit(PermCreateSecret); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		secret := Secret{
			ID:          data.RandIdent[ID](),
			Owners:      []user.ID{subject.ID()},
			LastMod:     time.Now(),
			Credentials: credentials,
		}

		optSec, err := secrets.FindByID(secret.ID)
		if err != nil {
			return "", fmt.Errorf("cannot find secret: %v", err)
		}

		if optSec.IsSome() {
			return "", fmt.Errorf("secret already exists")
		}

		return secret.ID, secrets.Save(secret)
	}
}
