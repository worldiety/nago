// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"iter"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
)

func NewListGlobalPermissions(rdb *rebac.DB) ListGlobalPermissions {
	return func(subject AuditableUser, uid ID) iter.Seq2[permission.ID, error] {
		return func(yield func(permission.ID, error) bool) {
			if subject.ID() != uid && !subject.HasPermission(PermFindByID) {
				yield("", PermissionDeniedError(PermFindByID))
				return
			}

			// this kind of query is constant to the number of permissions. the alternative is an
			// open search regarding all relations for the user, which may be an open set (like millions of entries)
			for perm := range permission.All() {
				ok, err := rdb.Contains(rebac.Triple{
					Source: rebac.Entity{
						Namespace: Namespace,
						Instance:  rebac.Instance(uid),
					},
					Relation: rebac.Relation(perm.ID),
					Target: rebac.Entity{
						Namespace: rebac.Global,
						Instance:  rebac.AllInstances,
					},
				})

				if err != nil {
					if !yield("", err) {
						return
					}

					continue
				}

				if ok {
					if !yield(perm.ID, nil) {
						return
					}
				}
			}
		}
	}
}
