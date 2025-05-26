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

func NewSetResourcePermissions(mutex *sync.Mutex, repo Repository) SetResourcePermissions {
	return func(subject AuditableUser, uid ID, resource Resource, permissions ...permission.ID) error {
		if err := subject.Audit(PermSetResourcePermissions); err != nil {
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
		slices.Sort(perms)
		slices.Sort(permissions)
		if slices.Equal(permissions, perms) {
			// nothing to write, exit early
			return nil
		}

		perms = slices.Compact(permissions)
		if usr.Resources == nil {
			usr.Resources = map[Resource][]permission.ID{}
		}
		usr.Resources[resource] = perms

		return repo.Save(usr)
	}
}
