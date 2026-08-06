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

// emailChangeOptions parametrizes [applyEmailChange] for the different callers.
type emailChangeOptions struct {
	// NotifyUser causes a verification mail to be sent to the new address, see the EMailChanged subscriber.
	NotifyUser bool

	// KeepVerified is set by the single sign on path, because the identity provider has already verified the
	// new address. Without it the user would be locked out, see NewAuthenticatesByPassword.
	KeepVerified bool
}

// applyEmailChange is the shared core of the administrative [ChangeOtherEmail] use case and the single sign on
// merge. The caller must already hold the global lock and must have done the audit, thus this must never be
// exported. It returns the updated user.
func applyEmailChange(bus events.Bus, repo Repository, idx *UserIndex, usr User, newEmail Email, opts emailChangeOptions) (User, error) {
	if !newEmail.Valid() {
		return usr, std.NewLocalizedError("Ungültige E-Mail-Adresse", "Die angegebene E-Mail-Adresse ist nicht gültig.").WithError(InvalidEMailErr)
	}

	if usr.Email.Equals(newEmail) {
		// nothing to do, especially do not invalidate an already verified mail address
		return usr, nil
	}

	// security note: the mail address is the login identity, thus it must be unique.
	inUse, err := idx.UsedByOther(newEmail, usr.ID)
	if err != nil {
		return usr, fmt.Errorf("cannot check mail uniqueness: %w", err)
	}

	if inUse {
		return usr, std.NewLocalizedError("Nutzer nicht aktualisiert", "Die E-Mail-Adresse wird bereits von einem anderen Nutzer verwendet.").WithError(EMailAlreadyInUseErr)
	}

	oldMail := usr.Email
	usr.Email = newEmail

	if !opts.KeepVerified {
		usr.EMailVerified = false
	}

	// security note: invalidate all codes, because they have been sent to the old address
	usr.VerificationCode = Code{}
	usr.PasswordRequestCode = Code{}

	if err := repo.Save(usr); err != nil {
		return usr, fmt.Errorf("cannot save user: %w", err)
	}

	// note: we do not create a verification code here, because the mail sending side creates a fresh
	// one anyway (see ResetVerificationCode).
	bus.Publish(EMailChanged{
		ID:         usr.ID,
		OldEMail:   oldMail,
		NewEMail:   newEmail,
		NotifyUser: opts.NotifyUser,
		ChangedAt:  time.Now(),
	})

	return usr, nil
}

func NewChangeOtherEmail(mutex *sync.Mutex, bus events.Bus, repo Repository, idx *UserIndex) ChangeOtherEmail {
	return func(subject AuditableUser, id ID, newEmail Email, notifyUser bool) error {
		if err := subject.Audit(PermChangeOtherEmail); err != nil {
			return err
		}

		// the mutex protects the uniqueness check inside applyEmailChange against concurrent modifications
		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return std.NewLocalizedError("Nutzer nicht aktualisiert", "Der Nutzer ist nicht (mehr) vorhanden.")
		}

		// note: we intentionally do NOT block SSO/NLS managed users here. If the mail address has been changed
		// within the identity provider, an admin must be able to fix the local address by hand, otherwise
		// MergeSingleSignOnUser cannot match the existing user anymore and would create a duplicate.
		_, err = applyEmailChange(bus, repo, idx, optUsr.Unwrap(), newEmail, emailChangeOptions{NotifyUser: notifyUser})

		return err
	}
}
