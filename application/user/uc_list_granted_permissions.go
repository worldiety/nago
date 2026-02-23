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
	"go.wdy.de/nago/application/rebac"
)

func NewListGrantedPermissions(rdb *rebac.DB) ListGrantedPermissions {
	return func(subject AuditableUser, id GrantingKey) ([]permission.ID, error) {

		res, uid := id.Split()
		if res.Name == "" {
			return nil, fmt.Errorf("invalid granting id")
		}

		// are we globally allowed?
		globalAllowed := subject.HasPermission(PermListGrantedPermissions)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasPermission(PermListGrantedPermissions)

		if !globalAllowed && !resAllowed {
			return nil, PermissionDeniedErr
		}

		var perms []permission.ID

		for triple, err := range rdb.Query(rebac.Select().Where().Source().Is(Namespace, rebac.Instance(uid))) {
			if err != nil {
				return nil, err
			}

			if triple.Target.Namespace == rebac.Namespace(res.Name) && triple.Target.Instance == rebac.Instance(res.ID) {
				perms = append(perms, permission.ID(triple.Relation))
			}
		}

		return perms, nil
	}
}
