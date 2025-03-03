package user

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"sync"
	"time"
)

func NewChangeOtherPassword(mutex *sync.Mutex, repo Repository) ChangeOtherPassword {
	return func(subject AuditableUser, uid ID, newPassword Password, newRepeated Password) error {
		if err := subject.Audit(PermChangeOtherPassword); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock() // this is really harsh and allows intentionally only to change one user per second

		if newPassword != newRepeated {
			return std.NewLocalizedError("Eingabefehler", "Das neue Kennwort unterscheidet sich in der wiederholten Eingabe.")
		}

		if err := newPassword.Validate(); err != nil {
			return err
		}

		// load existing user
		optUsr, err := repo.FindByID(uid)
		if err != nil {
			return fmt.Errorf("cannot find existing user: %w", err)
		}

		if optUsr.IsNone() {
			return fmt.Errorf("user has just disappeared")
		}

		usr := optUsr.Unwrap()

		// check if old password is the same as the new one
		if err := newPassword.CompareHashAndPassword(usr.Algorithm, usr.Salt, usr.PasswordHash); err == nil {
			return std.NewLocalizedError("Eingabebeschränkung", "Das alte Kennwort darf nicht identisch zum neuen Kennwort sein.")
		} else {
			// whatever, ignore any error (e.g. either discontinued hash or just a different password), thus continue and write over
		}

		// create new credentials
		newSalt, newHash, err := newPassword.Hash(Argon2IdMin)
		if err != nil {
			return err
		}

		usr.Salt = newSalt
		usr.PasswordHash = newHash
		usr.Algorithm = Argon2IdMin
		usr.LastPasswordChangedAt = time.Now()
		usr.RequirePasswordChange = false

		if err := repo.Save(usr); err != nil {
			return fmt.Errorf("cannot update user with new password: %w", err)
		}

		return nil
	}
}
