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

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

func NewUpdateMySecretGroups(mutex *sync.Mutex, bus events.Bus, repository Repository) UpdateMySecretGroups {
	return func(subject auth.Subject, id ID, groups []group.ID) error {
		if err := subject.Audit(PermUpdateMySecretGroups); err != nil {
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

		secret := optSecret.Unwrap()
		if !secret.HasAccess(subject) {
			return AccessDeniedErr
		}

		for _, gid := range groups {
			if slices.Contains(secret.Groups, gid) {
				// unchanged, keeping a group is always fine
				continue
			}

			if !subject.HasGroup(gid) {
				return std.NewLocalizedError("Secret Gruppen nicht aktualisiert", fmt.Sprintf("Die Gruppe '%v' kann nicht hinzugefügt werden, da das Konto selbst nicht zu der Gruppe gehört.", gid))
			}
		}

		// Groups in which the subject is not a member must survive, even if they have not been passed in.
		// Otherwise a subject would silently revoke the access of others just by saving a form which only
		// offers its own groups.
		var newGroups []group.ID
		for _, gid := range secret.Groups {
			if !subject.HasGroup(gid) {
				newGroups = append(newGroups, gid)
			}
		}

		for _, gid := range groups {
			if !slices.Contains(newGroups, gid) {
				newGroups = append(newGroups, gid)
			}
		}

		if !secret.IsOwner(subject) {
			// a subject which only has access through a group membership must not lock itself out
			probe := Secret{Owners: secret.Owners, Groups: newGroups}
			if !probe.HasAccess(subject) {
				return std.NewLocalizedError("Secret Gruppen nicht aktualisiert", "Die letzte Gruppe, über die der Zugriff auf dieses Secret besteht, kann nicht entfernt werden.")
			}
		}

		secret.Groups = newGroups

		if err := repository.Save(secret); err != nil {
			return fmt.Errorf("cannot save secret: %w", err)
		}

		bus.Publish(Updated{Secret: id})
		return nil
	}
}
