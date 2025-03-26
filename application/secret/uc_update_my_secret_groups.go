// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewUpdateMySecretGroups(mutex *sync.Mutex, repository Repository) UpdateMySecretGroups {
	return func(subject auth.Subject, id ID, groups []group.ID) error {
		if err := subject.Audit(PermUpdateMySecretGroups); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		for _, gid := range groups {
			if !subject.HasGroup(gid) {
				return std.NewLocalizedError("Secret Gruppen nicht aktualisiert", fmt.Sprintf("Die Gruppe '%v' kann nicht hinzugefügt werden, da das Konto selbst nicht zu der Gruppe gehört.", gid))
			}
		}

		optSecret, err := repository.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find secret: %w", err)
		}

		if optSecret.IsNone() {
			return std.NewLocalizedError("Secret Gruppen nicht aktualisiert", fmt.Sprintf("Das Secret existiert nicht: %v", id))
		}

		secret := optSecret.Unwrap()
		secret.Groups = groups

		return repository.Save(secret)
	}
}
