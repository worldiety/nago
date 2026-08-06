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

	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

func NewChangeOtherEmail(mutex *sync.Mutex, bus events.Bus, repo Repository, mailIdx *MailIndex) ChangeOtherEmail {
	return func(subject AuditableUser, id ID, newEmail Email, notifyUser bool) error {
		if err := subject.Audit(PermChangeOtherEmail); err != nil {
			return err
		}

		if !newEmail.Valid() {
			return std.NewLocalizedError("Ungültige E-Mail-Adresse", "Die angegebene E-Mail-Adresse ist nicht gültig.").WithError(InvalidEMailErr)
		}

		// the mutex protects the uniqueness check below against concurrent modifications
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
		if usr.Email.Equals(newEmail) {
			// nothing to do, especially do not invalidate an already verified mail address
			return nil
		}

		// security note: the mail address is the login identity, thus it must be unique.
		inUse, err := mailIdx.UsedByOther(newEmail, usr.ID)
		if err != nil {
			return fmt.Errorf("cannot check mail uniqueness: %w", err)
		}

		if inUse {
			return std.NewLocalizedError("Nutzer nicht aktualisiert", "Die E-Mail-Adresse wird bereits von einem anderen Nutzer verwendet.").WithError(EMailAlreadyInUseErr)
		}

		// note: we intentionally do NOT block SSO/NLS managed users here. If the mail address has been changed
		// within the identity provider, an admin must be able to fix the local address by hand, otherwise
		// MergeSingleSignOnUser cannot match the existing user anymore and would create a duplicate.
		oldMail := usr.Email
		usr.Email = newEmail
		usr.EMailVerified = false
		// security note: invalidate all codes, because they have been sent to the old address
		usr.VerificationCode = Code{}
		usr.PasswordRequestCode = Code{}

		if err := repo.Save(usr); err != nil {
			return fmt.Errorf("cannot save user: %w", err)
		}

		// note: we do not create a verification code here, because the mail sending side creates a fresh
		// one anyway (see ResetVerificationCode).
		bus.Publish(EMailChanged{
			ID:         usr.ID,
			OldEMail:   oldMail,
			NewEMail:   newEmail,
			NotifyUser: notifyUser,
			ChangedAt:  time.Now(),
		})

		return nil
	}
}
