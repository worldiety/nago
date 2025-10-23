// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"fmt"
	"slices"
	"sync"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

func NewUpdateMySecretOwners(mutex *sync.Mutex, bus events.Bus, repository Repository) UpdateMySecretOwners {
	return func(subject auth.Subject, id ID, users []user.ID) error {
		if err := subject.Audit(PermUpdateMySecretOwners); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optSecret, err := repository.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find secret: %w", err)
		}

		if optSecret.IsNone() {
			return std.NewLocalizedError("Secret Gruppen nicht aktualisiert", fmt.Sprintf("Das Secret existiert nicht: %v", id))
		}

		if !slices.Contains(users, subject.ID()) {
			// we don't allow the mistake to remove our own ownership. If required, someone else must do that
			// otherwise we may cause orphaned secrets which can never be recovered, well if the user is removed
			// we still have that problem
			users = append(users, subject.ID())
		}

		secret := optSecret.Unwrap()
		secret.Owners = users

		if err := repository.Save(secret); err != nil {
			return fmt.Errorf("cannot save secret: %w", err)
		}

		bus.Publish(Updated{Secret: id})
		return nil
	}
}
