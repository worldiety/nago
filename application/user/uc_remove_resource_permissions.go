// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
)

func NewRemoveResourcePermissions(rdb *rebac.DB) RemoveResourcePermissions {
	return func(subject AuditableUser, uid ID, resource Resource, permissions ...permission.ID) error {
		if err := subject.Audit(PermRemoveResourcePermissions); err != nil {
			return err
		}

		for _, id := range permissions {
			err := rdb.Delete(rebac.Triple{
				Source: rebac.Entity{
					Namespace: Namespace,
					Instance:  rebac.Instance(uid),
				},
				Relation: rebac.Relation(id),
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
}
