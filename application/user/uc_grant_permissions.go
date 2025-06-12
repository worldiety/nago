// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"os"
	"slices"
	"sync"
)

func NewGrantPermissions(mutex *sync.Mutex, repo Repository, index GrantingIndexRepository, findUserByID FindByID) GrantPermissions {
	return func(subject AuditableUser, id GrantingKey, permissions ...permission.ID) error {
		res, uid := id.Split()
		if res.Name == "" {
			return fmt.Errorf("invalid granting id")
		}

		// are we globally allowed?
		globalAllowed := subject.HasResourcePermission(index.Name(), string(id), PermGrantPermissions)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasResourcePermission(res.Name, res.ID, PermGrantPermissions)

		if !globalAllowed && !resAllowed {
			return PermissionDeniedErr
		}

		// security note: our permissions are checked above
		optUsr, err := findUserByID(SU(), uid)
		if err != nil {
			return err
		}

		if optUsr.IsNone() {
			return fmt.Errorf("user not found: %w", os.ErrNotExist)
		}

		optGrant, err := index.FindByID(id)
		if err != nil {
			return err
		}

		slices.Sort(permissions)

		// security note: we checked above with a different rule set
		if err := setResourcePermissions(mutex, repo, uid, res, permissions...); err != nil {
			return fmt.Errorf("cannot set user resource permission: %w", err)
		}

		if len(permissions) == 0 {
			if err := index.DeleteByID(id); err != nil {
				return fmt.Errorf("failed to delete grant from index: %w", err)
			}

			return nil
		}

		if optGrant.IsNone() {
			// only write into index, if actually required
			return index.Save(Granting{ID: id})
		}

		// index has already a grant, nothing to do
		return nil
	}
}

func setResourcePermissions(mutex *sync.Mutex, repo Repository, uid ID, resource Resource, permissions ...permission.ID) error {

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
