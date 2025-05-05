// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/permission"
	"os"
	"slices"
	"sync"
)

func NewAddResourcePermissions(mutex *sync.Mutex, repo Repository) AddResourcePermissions {
	return func(subject AuditableUser, uid ID, resource Resource, permissions ...permission.ID) error {
		if err := subject.Audit(PermAddResourcePermissions); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := repo.FindByID(uid)
		if err != nil {
			return err
		}

		if optUsr.IsNone() {
			return os.ErrNotExist
		}

		usr := optUsr.Unwrap()
		perms := usr.Resources[resource]
		changed := false
		for _, newPerm := range permissions {
			if !slices.Contains(perms, newPerm) {
				perms = append(perms, newPerm)
				changed = true
			}
		}

		if !changed {
			// nothing to write, exit early
			return nil
		}

		slices.Sort(perms)
		perms = slices.Compact(perms)
		usr.Resources[resource] = perms

		return repo.Save(usr)
	}
}
