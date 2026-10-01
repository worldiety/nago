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

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewUpdateVerification(mutex *sync.Mutex, repo Repository) UpdateVerification {
	return func(subject permission.Auditable, id ID, verified bool) error {
		if err := subject.Audit(PermUpdateOtherContact); err != nil {
			return err
		}

		// mutex is important, otherwise we may re-create a user accidentally
		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return std.NewLocalizedError("Nutzer nicht aktualisiert", "Der Nutzer ist nicht (mehr) vorhanden.")
		}

		usr := optUsr.Unwrap()
		setVerified(&usr, verified)
		return repo.Save(usr)
	}
}

// setVerified applies the verification state. A verified user needs no code anymore. An unverified user keeps a
// still valid code or gets a fresh one, so that the verification mail can be sent again.
func setVerified(usr *User, verified bool) {
	usr.EMailVerified = verified
	if verified {
		usr.VerificationCode = Code{}
		return
	}

	if usr.VerificationCode.Value == "" || time.Now().After(usr.VerificationCode.ValidUntil) {
		usr.VerificationCode = NewCode(DefaultVerificationLifeTime)
	}
}
