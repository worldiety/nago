// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/consent"
	"os"
	"sync"
	"time"
)

func NewConsent(mutex *sync.Mutex, usersRepo Repository) Consent {
	return func(subject AuditableUser, uid ID, cid consent.ID, status consent.Status) error {
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
		updated := false
		for idx, c := range usr.Consents {
			if c.ID == cid {
				c.History = append(c.History, consent.Action{
					At:     time.Now(),
					Status: status,
				})

				usr.Consents[idx] = c
				updated = true

				break
			}
		}

		if !updated {
			usr.Consents = append(usr.Consents, consent.Consent{
				ID: cid,
				History: []consent.Action{
					{At: time.Now(), Status: status},
				},
			})
		}

		return usersRepo.Save(usr)
	}
}
