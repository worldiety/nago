// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
)

func NewUpdatePermissions(rdb *rebac.DB) UpdatePermissions {
	return func(subject permission.Auditable, id ID, permissions []permission.ID) error {
		if err := subject.Audit(PermUpdate); err != nil {
			return err
		}
		
		// remove all perms
		for perm := range permission.All() {
			err := rdb.Delete(rebac.Triple{
				Source: rebac.Entity{
					Namespace: Namespace,
					Instance:  rebac.Instance(id),
				},
				Relation: rebac.Relation(perm.ID),
				Target: rebac.Entity{
					Namespace: rebac.Global,
					Instance:  rebac.AllInstances,
				},
			})
			if err != nil {
				return err
			}
		}

		// than just store new relations
		for _, pid := range permissions {
			err := rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: Namespace,
					Instance:  rebac.Instance(id),
				},
				Relation: rebac.Relation(pid),
				Target: rebac.Entity{
					Namespace: rebac.Global,
					Instance:  rebac.AllInstances,
				},
			})
			if err != nil {
				return err
			}
		}

		return nil
	}
}
