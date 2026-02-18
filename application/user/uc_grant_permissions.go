// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"os"
	"sync"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
)

func NewGrantPermissions(mutex *sync.Mutex, repo Repository, findUserByID FindByID, rdb *rebac.DB) GrantPermissions {
	return func(subject AuditableUser, id GrantingKey, permissions ...permission.ID) error {
		res, uid := id.Split()
		if res.Name == "" {
			return fmt.Errorf("invalid granting id")
		}

		// are we globally allowed?
		globalAllowed := subject.HasPermission(PermGrantPermissions)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasPermission(PermGrantPermissions)

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

		// security note: we checked above with a different rule set
		if err := setResourcePermissions(rdb, uid, res, permissions...); err != nil {
			return fmt.Errorf("cannot set user resource permission: %w", err)
		}

		// index has already a grant, nothing to do
		return nil
	}
}

func setResourcePermissions(rdb *rebac.DB, uid ID, resource Resource, permissions ...permission.ID) error {

	// delete all possible permissions for this resource, note that unknown perms are still kept
	for perm := range permission.All() {
		err := rdb.Delete(rebac.Triple{
			Source: rebac.Entity{
				Namespace: Namespace,
				Instance:  rebac.Instance(uid),
			},
			Relation: rebac.Relation(perm.ID),
			Target: rebac.Entity{
				Namespace: rebac.Namespace(resource.Name),
				Instance:  rebac.Instance(resource.ID),
			},
		})

		if err != nil {
			return err
		}
	}

	// insert only the specified back
	for _, pid := range permissions {
		err := rdb.Put(rebac.Triple{
			Source: rebac.Entity{
				Namespace: Namespace,
				Instance:  rebac.Instance(uid),
			},
			Relation: rebac.Relation(pid),
			Target: rebac.Entity{
				Namespace: rebac.Namespace(resource.Name),
				Instance:  rebac.Instance(resource.ID),
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
}
