// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"iter"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
)

func NewListRoles(rdb *rebac.DB) ListRoles {
	return func(subject AuditableUser, uid ID) iter.Seq2[role.ID, error] {
		return func(yield func(role.ID, error) bool) {
			if subject.ID() != uid && !subject.HasPermission(PermFindByID) {
				yield("", PermissionDeniedError(PermFindByID))
				return
			}

			ListRolesFrom(rdb, uid)(yield)
		}
	}
}

func ListRolesFrom(rdb *rebac.DB, uid ID) iter.Seq2[role.ID, error] {
	return func(yield func(role.ID, error) bool) {
		it := rdb.Query(rebac.Select().
			Where().Source().IsNamespace(role.Namespace).
			Where().Relation().Has(rebac.Member).
			Where().Target().Is(Namespace, rebac.Instance(uid)),
		)

		for triple, err := range it {
			if err != nil {
				if !yield("", err) {
					return
				}

				continue
			}

			if !yield(role.ID(triple.Source.Instance), nil) {
				return
			}
		}
	}
}
