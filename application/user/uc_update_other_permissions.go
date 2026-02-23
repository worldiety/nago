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

func NewUpdateOtherPermissions(rdb *rebac.DB) UpdateOtherPermissions {
	return func(subject AuditableUser, id ID, permissions []permission.ID) error {
		if err := subject.Audit(PermUpdateOtherPermissions); err != nil {
			return err
		}

		// first remove all existing permissions to keep the correct semantics of this use case
		// this is not exactly the same, because if permissions were once declared but are currently
		// unknown, the will be kept "forever". We may mitigate that using a kind of garbage collection in rdb.
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
