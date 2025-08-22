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

	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

func NewUpdateOtherContact(mutex *sync.Mutex, bus events.Bus, repo Repository) UpdateOtherContact {
	return func(subject AuditableUser, id ID, contact Contact) error {
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
		usr.Contact = contact

		if err := repo.Save(usr); err != nil {
			return fmt.Errorf("cannot save user: %w", err)
		}

		bus.Publish(ContactUpdated{
			ID:      subject.ID(),
			Contact: contact,
		})

		return nil
	}
}
