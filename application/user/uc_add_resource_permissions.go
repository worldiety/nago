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

func NewAddResourcePermissions(rdb *rebac.DB) AddResourcePermissions {
	return func(subject AuditableUser, uid ID, resource Resource, permissions ...permission.ID) error {
		if err := subject.Audit(PermAddResourcePermissions); err != nil {
			return err
		}

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
}
