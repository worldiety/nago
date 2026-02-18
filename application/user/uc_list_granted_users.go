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
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xmaps"
)

func NewListGrantedUsers(rdb *rebac.DB) ListGrantedUsers {
	return func(subject AuditableUser, res Resource) iter.Seq2[ID, error] {
		//myID := NewGrantingKey(res, subject.ID())

		// are we globally allowed?
		globalAllowed := subject.HasPermission(PermListGrantedUsers)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasPermission(PermListGrantedUsers)

		if !globalAllowed && !resAllowed {
			return xiter.WithError[ID](PermissionDeniedErr)
		}

		return func(yield func(ID, error) bool) {
			tmp := map[ID]struct{}{}

			it := rdb.Query(rebac.Select().Where().Target().Is(rebac.Namespace(res.Name), rebac.Instance(res.ID)))

			// a single user may have multiple relations for a single resource, this use case is weired
			for triple, err := range it {
				if err != nil {
					yield("", err)
					return
				}
				tmp[ID(triple.Source.Instance)] = struct{}{}
			}

			for _, id := range xmaps.SortedKeys(tmp) {
				if !yield(id, nil) {
					return
				}
			}
		}
	}
}
