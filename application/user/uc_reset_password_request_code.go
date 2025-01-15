package user

import (
	"fmt"
	"sync"
	"time"
)

func NewResetPasswordRequestCode(mutex *sync.Mutex, sysUser SysUser, repository Repository, byMail FindByMail) ResetPasswordRequestCode {
	return func(mail Email, lifetime time.Duration) (code string, err error) {
		mutex.Lock()
		defer mutex.Unlock()

		optUser, err := byMail(sysUser(), mail)
		if err != nil {
			return "", fmt.Errorf("cannot find user by mail: %w", err)
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

		user.PasswordRequestCode = NewCode(lifetime)

		return user.PasswordRequestCode.Value, repository.Save(user)
	}
}
