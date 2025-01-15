package user

import (
	"fmt"
	"sync"
	"time"
)

func NewResetVerificationCode(mutex *sync.Mutex, repository Repository) ResetVerificationCode {
	return func(id ID, lifetime time.Duration) (code string, err error) {
		mutex.Lock()
		defer mutex.Unlock()

		optUser, err := repository.FindByID(id)
		if err != nil {
			return "", fmt.Errorf("cannot find user: %w", err)
		}

		if optUser.IsNone() {
			// security note: do not expose a readable message here
			return "", fmt.Errorf("user is none")
		}

		user := optUser.Unwrap()
		if !user.Enabled() {
			// security note: do not expose a readable message here
			return "", fmt.Errorf("user is disabled")
		}

		user.VerificationCode = NewCode(lifetime)

		return user.VerificationCode.Value, repository.Save(user)
	}
}
