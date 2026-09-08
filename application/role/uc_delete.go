// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"sync"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/events"
)

func NewDelete(mutex *sync.Mutex, repo Repository, bus events.Bus, rdb *rebac.DB) Delete {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDelete); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		// A system role is declared by a module and recreated on the next start, so deleting it would only
		// produce a confusing gap rather than the intended effect.
		optRole, err := repo.FindByID(id)
		if err != nil {
			return err
		}

		if optRole.IsSome() && optRole.Unwrap().IsSystem() {
			return errSystemRoleProtected(id)
		}

		if err := repo.DeleteByID(id); err != nil {
			return err
		}

		// Purge every triple this role takes part in, on both sides of the relation.
		//
		// Both directions are needed and the source side is the one that actually carries the role's power:
		// a permission grant is stored as (role -> permission -> global) by [NewUpdatePermissions] and a
		// membership as (role -> member -> user), so the role is the source in both cases. Purging only the
		// target side left those triples behind as orphans - and because role ids may be chosen by the
		// operator, a role recreated under a previously used id silently inherited the permissions and the
		// members of its namesake.
		instance := rebac.Instance(id)
		if err := rdb.DeleteByQuery(rebac.Select().Where().Source().Is(Namespace, instance)); err != nil {
			return err
		}

		if err := rdb.DeleteByQuery(rebac.Select().Where().Target().Is(Namespace, instance)); err != nil {
			return err
		}

		bus.Publish(Deleted{Role: id})
		return nil
	}
}
