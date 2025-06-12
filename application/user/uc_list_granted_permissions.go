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
)

func NewListGrantedPermissions(repo Repository, findUserByID FindByID) ListGrantedPermissions {
	return func(subject AuditableUser, id GrantingKey) ([]permission.ID, error) {

		res, uid := id.Split()
		if res.Name == "" {
			return nil, fmt.Errorf("invalid granting id")
		}

		// are we globally allowed?
		globalAllowed := subject.HasResourcePermission(repo.Name(), string(id), PermListGrantedPermissions)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasResourcePermission(res.Name, res.ID, PermListGrantedPermissions)

		if !globalAllowed && !resAllowed {
			return nil, PermissionDeniedErr
		}

		// security note: our permissions have been checked above
		optUsr, err := findUserByID(SU(), uid)
		if err != nil {
			return nil, err
		}

		if optUsr.IsNone() {
			// not sure, what the best behavior is here
			return nil, nil
		}

		usr := optUsr.Unwrap()
		perms, ok := usr.Resources[res]
		if !ok {
			return nil, nil
		}

		return perms, nil
	}
}
