// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/permission"
	"slices"
	"sync"
)

func NewRemoveResourcePermissions(mutex *sync.Mutex, repo Repository) RemoveResourcePermissions {
	return func(subject AuditableUser, uid ID, resource Resource, permissions ...permission.ID) error {
		if err := subject.Audit(PermRemoveResourcePermissions); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := repo.FindByID(uid)
		if err != nil {
			return err
		}

		if optUsr.IsNone() {
			// that actually means that all permissions are gone
			return nil
		}

		usr := optUsr.Unwrap()
		perms := usr.Resources[resource]
		changed := false

		perms = slices.DeleteFunc(perms, func(id permission.ID) bool {
			toRemove := slices.Contains(permissions, id)
			if toRemove {
				changed = true
			}

			return toRemove
		})

		if !changed {
			// nothing to write, exit early
			return nil
		}

		// clean up properly
		slices.Sort(perms)
		perms = slices.Compact(perms)
		if usr.Resources == nil {
			usr.Resources = map[Resource][]permission.ID{}
		}
		
		usr.Resources[resource] = perms

		return repo.Save(usr)
	}
}
