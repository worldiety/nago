// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"context"
	"time"

	"go.wdy.de/nago/application/migration"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
)

type migrateRolePermsToReBAC struct {
	roleStore blob.Store
	rdb       *rebac.DB
}

func newMigrateRolePermsToReBAC(store blob.Store, rdb *rebac.DB) migrateRolePermsToReBAC {
	return migrateRolePermsToReBAC{roleStore: store, rdb: rdb}
}

func (m migrateRolePermsToReBAC) Version() migration.Version {
	return migration.NewVersion(2026, time.January, 17, 17, 46, "MoveRolePermsToReBAC")
}

func (m migrateRolePermsToReBAC) Migrate(ctx context.Context) error {
	// legacyRole is the last data model that still contains permissions
	type legacyRole struct {
		ID          role.ID         `json:"id,omitempty"`
		Permissions []permission.ID `json:"permissions,omitempty"`
	}

	for rid, err := range m.roleStore.List(ctx, blob.ListOptions{}) {
		if err != nil {
			return err
		}
		optRole, err := json.Get[legacyRole](m.roleStore, rid)
		if err != nil {
			return err
		}

		if optRole.IsNone() {
			continue
		}

		usr := optRole.Unwrap()

		// migrate direct and global permissions
		for _, id := range usr.Permissions {
			err := m.rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: role.Namespace,
					Instance:  rebac.Instance(rid),
				},
				Relation: rebac.Relation(id),
				Target: rebac.Entity{
					Namespace: rebac.Global,
					Instance:  rebac.AllInstances,
				},
			})
			if err != nil {
				return err
			}
		}

	}

	return nil
}
