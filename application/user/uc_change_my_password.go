package user

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewChangeMyPassword(mutex *sync.Mutex, repo Repository) ChangeMyPassword {
	return func(subject AuditableUser, oldPassword, newPassword, newRepeated Password) error {
		mutex.Lock()
		defer mutex.Unlock() // this is really harsh and allows intentionally only to change one user per second

		if !subject.Valid() {
			return fmt.Errorf("invalid subject")
		}

		if oldPassword == newPassword {
			return std.NewLocalizedError("Eingabebeschränkung", "Das alte und das neue Kennwort müssen sich unterscheiden.")
		}

		if newPassword != newRepeated {
			return std.NewLocalizedError("Eingabefehler", "Das neue Kennwort unterscheidet sich in der wiederholten Eingabe.")
		}

		if err := newPassword.Validate(); err != nil {
			return err
		}

		// check if old password authenticates
		optUsr, err := repo.FindByID(subject.ID())
		if err != nil {
			return fmt.Errorf("cannot find existing user: %w", err)
		}

		if optUsr.IsNone() {
			return fmt.Errorf("user has just disappeared")
		}

		usr := optUsr.Unwrap()

		if err := oldPassword.CompareHashAndPassword(Argon2IdMin, usr.Salt, usr.PasswordHash); err != nil {
			return err
		}

		// create new credentials
		newSalt, newHash, err := newPassword.Hash(Argon2IdMin)
		if err != nil {
			return err
		}

		usr.Salt = newSalt
		usr.PasswordHash = newHash
		usr.Algorithm = Argon2IdMin

		if err := repo.Save(usr); err != nil {
			return fmt.Errorf("cannot update user with new password: %w", err)
		}

		return nil
	}
}
