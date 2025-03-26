// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/pkg/std"
	"sync"
	"time"
)

func NewChangePasswordWithCode(mutex *sync.Mutex, su SysUser, repo Repository, changePwd ChangeOtherPassword) ChangePasswordWithCode {
	return func(uid ID, code string, newPassword Password, newRepeated Password) error {
		optUser, err := repo.FindByID(uid)
		if err != nil {
			return err
		}

		accountErr := std.NewLocalizedError("Kennwortänderung", "Das Konto existiert nicht, ist deaktiviert oder der Code ist bereits abgelaufen.").WithError(AccountVerificationFailed)
		if optUser.IsNone() {
			// security note: don't expose any detail
			return accountErr
		}

		user := optUser.Unwrap()
		if user.PasswordRequestCode.ValidUntil.Before(time.Now()) {
			return accountErr
		}

		if len(user.PasswordRequestCode.Value) < 6 {
			// security note: don't fool ourselves
			return accountErr
		}

		if user.PasswordRequestCode.Value != code {
			return accountErr
		}

		if !user.Enabled() {
			return accountErr
		}

		if err := changePwd(su(), uid, newPassword, newRepeated); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		// double-check to mitigate races and inserting by accident when invalidating code
		optUser, err = repo.FindByID(uid)
		if err != nil {
			return err
		}

		if optUser.IsNone() {
			return accountErr
		}

		user = optUser.Unwrap()
		user.PasswordRequestCode = Code{}
		user.VerificationCode = Code{}
		// security note: it is ok, to verify and remove any need for the verification code,
		// if the password has been reset by code which was also by mail; likely, just as [ConfirmMail].
		user.EMailVerified = true

		return repo.Save(user)
	}
}
