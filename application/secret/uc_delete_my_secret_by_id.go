package secret

import (
	"go.wdy.de/nago/auth"
)

func NewDeleteMySecretByID(repository Repository) DeleteMySecretByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteMySecretByID); err != nil {
			return err
		}

		return repository.DeleteByID(id)
	}
}
