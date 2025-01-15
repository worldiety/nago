package user

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"sync"
	"time"
)

const AccountVerificationFailed std.Error = "account verification failed"

func NewConfirmMail(mutex *sync.Mutex, repository Repository) ConfirmMail {
	return func(id ID, code string) error {
		mutex.Lock()
		defer mutex.Unlock()

		optUser, err := repository.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find user: %w", err)
		}

		accountErr := std.NewLocalizedError("Kontoverifikation fehlgeschlagen", "Das Konto existiert nicht, ist deaktiviert oder der Code ist bereits abgelaufen.").WithError(AccountVerificationFailed)
		if optUser.IsNone() {
			return accountErr
		}

		user := optUser.Unwrap()
		if user.VerificationCode.ValidUntil.Before(time.Now()) {
			return accountErr
		}

		if len(user.VerificationCode.Value) < 6 {
			// security note: don't fool ourselves
			return accountErr
		}

		if user.VerificationCode.Value != code {
			return accountErr
		}

		if !user.Enabled() {
			return accountErr
		}

		user.EMailVerified = true
		user.VerificationCode = Code{}

		return repository.Save(user)
	}
}
