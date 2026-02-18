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

func NewAddUserToGroup(rdb *rebac.DB) AddUserToGroup {
	return func(subject AuditableUser, id ID, gid group.ID) error {
		if err := subject.Audit(PermUpdateOtherGroups); err != nil {
			return err
		}

		return rdb.Put(rebac.Triple{
			Source: rebac.Entity{
				Namespace: Namespace,
				Instance:  rebac.Instance(id),
			},
			Relation: rebac.Member,
			Target: rebac.Entity{
				Namespace: group.Namespace,
				Instance:  rebac.Instance(gid),
			},
		})
	}
}
