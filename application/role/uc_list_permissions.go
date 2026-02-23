// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"iter"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
)

func NewListPermissions(rdb *rebac.DB) ListPermissions {
	return func(subject permission.Auditable, id ID) iter.Seq2[permission.ID, error] {
		return func(yield func(permission.ID, error) bool) {
			if err := subject.Audit(PermFindByID); err != nil {
				yield("", err)
				return
			}

			ListPermissionsFrom(rdb, id)(yield)
		}
	}
}

func ListPermissionsFrom(rdb *rebac.DB, rid ID) iter.Seq2[permission.ID, error] {
	return func(yield func(permission.ID, error) bool) {
		it := rdb.Query(rebac.Select().
			Where().Source().Is(Namespace, rebac.Instance(rid)).
			// relation == ?
			Where().Target().Is(rebac.Global, rebac.AllInstances),
		)

		for triple, err := range it {
			if !yield(permission.ID(triple.Relation), err) {
				return
			}
		}
	}
}
