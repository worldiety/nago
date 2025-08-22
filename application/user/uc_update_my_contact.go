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

func NewUpdateMyContact(mutex *sync.Mutex, bus events.Bus, repo Repository) UpdateMyContact {
	return func(subject AuditableUser, contact Contact) error {
		if !subject.Valid() {
			// bootstrap error message
			return std.NewLocalizedError("Nicht eingeloggt", "Diese Funktion steht nur eingeloggten Nutzern zur Verfügung.")
		}

		// mutex is important, otherwise we may re-create a user accidentally
		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := repo.FindByID(subject.ID())
		if err != nil {
			return fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return std.NewLocalizedError("Nicht eingeloggt", "Der Nutzer ist nicht (mehr) vorhanden.")
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
