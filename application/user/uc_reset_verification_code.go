// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"sync"
	"time"
)

// DefaultVerificationLifeTime is currently 48 hours. We got so many usage reports, that mails do not arrive in time.
// Sometimes multiple hours to late, probably due to grey listing or similar anti-spam techniques.
const DefaultVerificationLifeTime = time.Hour * 48

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
