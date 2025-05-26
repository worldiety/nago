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
)

func NewListGrants(repo Repository, findUserByID user.FindByID) ListGrants {
	return func(subject auth.Subject, id ID) ([]permission.ID, error) {

		res, uid := id.Split()
		if res.Name == "" {
			return nil, fmt.Errorf("invalid granting id")
		}

		// are we globally allowed?
		globalAllowed := subject.HasResourcePermission(repo.Name(), string(id), PermGrant)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasResourcePermission(res.Name, res.ID, PermGrant)

		if !globalAllowed && !resAllowed {
			return nil, user.PermissionDeniedErr
		}

		// security note: our permissions have been checked above
		optUsr, err := findUserByID(user.SU(), uid)
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
