// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
)

func NewUpdateOtherRoles(rdb *rebac.DB) UpdateOtherRoles {
	return func(subject AuditableUser, id ID, roles []role.ID) error {
		if err := subject.Audit(PermUpdateOtherContact); err != nil {
			return err
		}

		// first remove all existing roles to keep the correct semantics of this use case

		err := rdb.DeleteByQuery(rebac.Select().
			Where().Source().IsNamespace(role.Namespace).
			Where().Relation().Has(rebac.Member).
			Where().Target().Is(Namespace, rebac.Instance(id)))

		if err != nil {
			return err
		}

		// than just store new relations
		for _, rid := range roles {
			err := rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: role.Namespace,
					Instance:  rebac.Instance(rid),
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
