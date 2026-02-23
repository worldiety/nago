// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"iter"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
)

func NewListGroups(rdb *rebac.DB) ListGroups {
	return func(subject AuditableUser, uid ID) iter.Seq2[group.ID, error] {
		return func(yield func(group.ID, error) bool) {
			if subject.ID() != uid && !subject.HasPermission(PermFindByID) {
				yield("", PermissionDeniedError(PermFindByID))
				return
			}

			ListGroupsFrom(rdb, uid)(yield)
		}
	}
}

func ListGroupsFrom(rdb *rebac.DB, uid ID) iter.Seq2[group.ID, error] {
	return func(yield func(group.ID, error) bool) {
		it := rdb.Query(rebac.Select().
			Where().Source().IsNamespace(group.Namespace).
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

			if !yield(group.ID(triple.Source.Instance), nil) {
				return
			}
		}
	}
}
