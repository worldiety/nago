// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

func NewChangeMyPassword(mutex *sync.Mutex, repo Repository) ChangeMyPassword {
	return func(subject AuditableUser, oldPassword, newPassword, newRepeated Password) error {
		mutex.Lock()
		defer mutex.Unlock() // this is really harsh and allows intentionally only to change one user per second

		if !subject.Valid() {
			return fmt.Errorf("invalid subject")
		}

		if oldPassword == newPassword {
			return NewPasswordMustBeDifferentFromOldPasswordErr
		}

		if newPassword != newRepeated {
			return PasswordsDontMatchErr
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
			if errors.Is(err, noLoginErr) {
				return InvalidOldPasswordErr
			}
			
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
		usr.LastPasswordChangedAt = time.Now()
		usr.RequirePasswordChange = false

		if err := repo.Save(usr); err != nil {
			return fmt.Errorf("cannot update user with new password: %w", err)
		}

		return nil
	}
}
