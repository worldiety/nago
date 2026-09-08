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

func NewUpsertPermissions(rdb *rebac.DB) UpsertPermissions {
	return func(subject permission.Auditable, id ID, permissions []permission.ID) error {
		if err := subject.Audit(PermUpdate); err != nil {
			return err
		}

		if len(permissions) == 0 {
			return nil
		}

		// Read the current grants first so an already satisfied role costs a single query and no write at
		// all. This is what makes the use case safe to call on every start-up: the alternative,
		// [NewUpdatePermissions], issues one delete per *declared* permission in the system before writing
		// anything, which would turn each boot into an avoidable write storm.
		granted := make(map[permission.ID]bool)
		for pid, err := range ListPermissionsFrom(rdb, id) {
			if err != nil {
				return err
			}

			granted[pid] = true
		}

		source := rebac.Entity{Namespace: Namespace, Instance: rebac.Instance(id)}
		target := rebac.Entity{Namespace: rebac.Global, Instance: rebac.AllInstances}

		for _, pid := range permissions {
			if granted[pid] {
				continue
			}

			if err := rdb.Put(rebac.Triple{
				Source:   source,
				Relation: rebac.Relation(pid),
				Target:   target,
			}); err != nil {
				return err
			}

			// A duplicate in the argument must not cause a second write.
			granted[pid] = true
		}

		return nil
	}
}
