// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
)

func NewUpdateOtherGroups(rdb *rebac.DB, allGroups group.FindAll) UpdateOtherGroups {
	return func(subject AuditableUser, id ID, groups []group.ID) error {
		if err := subject.Audit(PermUpdateOtherGroups); err != nil {
			return err
		}

		// first remove all existing groups to keep the correct semantics of this use case
		for grp, err := range allGroups(SU()) {
			if err != nil {
				return err
			}

			err := rdb.Delete(rebac.Triple{
				Source: rebac.Entity{
					Namespace: group.Namespace,
					Instance:  rebac.Instance(grp.ID),
				},
				Relation: rebac.Member,
				Target: rebac.Entity{
					Namespace: Namespace,
					Instance:  rebac.Instance(id),
				},
			})
			if err != nil {
				return err
			}
		}

		// than just store new relations
		for _, gid := range groups {
			err := rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: group.Namespace,
					Instance:  rebac.Instance(gid),
				},
				Relation: rebac.Member,
				Target: rebac.Entity{
					Namespace: Namespace,
					Instance:  rebac.Instance(id),
				},
			})
			if err != nil {
				return err
			}
		}

		return nil
	}
}
