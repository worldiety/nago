// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package grant

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"os"
	"slices"
)

func NewGrant(repo Repository, findUserByID user.FindByID, setUserPerm user.SetResourcePermissions) Grant {
	return func(subject auth.Subject, id ID, permissions ...permission.ID) error {
		res, uid := id.Split()
		if res.Name == "" {
			return fmt.Errorf("invalid granting id")
		}

		// are we globally allowed?
		globalAllowed := subject.HasResourcePermission(repo.Name(), string(id), PermGrant)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasResourcePermission(res.Name, res.ID, PermGrant)

		if !globalAllowed && !resAllowed {
			return user.PermissionDeniedErr
		}

		// security note: our permissions are checked above
		optUsr, err := findUserByID(user.SU(), uid)
		if err != nil {
			return err
		}

		if optUsr.IsNone() {
			return fmt.Errorf("user not found: %w", os.ErrNotExist)
		}

		optGrant, err := repo.FindByID(id)
		if err != nil {
			return err
		}

		slices.Sort(permissions)

		// security note: we checked above with a different rule set
		if err := setUserPerm(user.SU(), uid, res, permissions...); err != nil {
			return fmt.Errorf("cannot set user resource permission: %w", err)
		}

		if len(permissions) == 0 {
			if err := repo.DeleteByID(id); err != nil {
				return fmt.Errorf("failed to delete grant from index: %w", err)
			}

			return nil
		}

		if optGrant.IsNone() {
			// only write into index, if actually required
			return repo.Save(Granting{ID: id})
		}

		// index has already a grant, nothing to do
		return nil
	}
}
