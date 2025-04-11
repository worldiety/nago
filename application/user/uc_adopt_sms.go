// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"os"
	"sync"
	"time"
)

func NewAdoptSMS(mutex *sync.Mutex, usersRepo Repository) AdoptSMS {
	return func(subject AuditableUser, uid ID, adopt bool) error {
		mutex.Lock()
		defer mutex.Unlock()

		self := subject.ID() == uid
		if !self {
			// it is not allowed - even for admins - to enable this, because at any time we may get sued
			// badly for any small error. Thus, just be sure to not technically allow that at all.
			return PermissionDeniedErr
		}

		optUsr, err := usersRepo.FindByID(uid)
		if err != nil {
			return err
		}

		if optUsr.IsNone() {
			return os.ErrNotExist
		}

		usr := optUsr.Unwrap()
		if adopt {
			usr.SMS = LegalAdoption{
				ApprovedAt: time.Now(),
				Name:       "SMS",
			}
		} else {
			usr.SMS = LegalAdoption{}
		}

		return usersRepo.Save(usr)
	}
}
