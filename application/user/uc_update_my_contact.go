package user

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
)

func NewUpdateMyContact(repo Repository) UpdateMyContact {
	return func(subject AuditableUser, contact Contact) error {
		if !subject.Valid() {
			// bootstrap error message
			return std.NewLocalizedError("Nicht eingeloggt", "Diese Funktion steht nur eingeloggten Nutzern zur Verfügung.")
		}

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
