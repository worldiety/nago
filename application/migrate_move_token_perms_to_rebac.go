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

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/migration"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/token"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
)

type migrateTokenPermsToReBAC struct {
	tokenStore blob.Store
	rdb        *rebac.DB
}

func newMigrateTokenPermsToReBAC(store blob.Store, rdb *rebac.DB) migrateTokenPermsToReBAC {
	return migrateTokenPermsToReBAC{tokenStore: store, rdb: rdb}
}

func (m migrateTokenPermsToReBAC) Version() migration.Version {
	return migration.NewVersion(2026, time.January, 16, 14, 41, "MoveTokenPermsToReBAC")
}

func (m migrateTokenPermsToReBAC) Migrate(ctx context.Context) error {
	// legacyToken is the last data model that still contains roles, groups, permissions and resources
	type legacyToken struct {
		ID token.ID `json:"id"`

		// Other permissions rules
		Groups      []group.ID                             `json:"groups,omitempty"`
		Roles       []role.ID                              `json:"roles,omitempty"`
		Permissions []permission.ID                        `json:"permissions,omitempty"`
		Resources   map[legacyUserResource][]permission.ID `json:"resources,omitempty" json:"resources,omitempty"`
	}

	for uid, err := range m.tokenStore.List(ctx, blob.ListOptions{}) {
		if err != nil {
			return err
		}
		optUsr, err := json.Get[legacyToken](m.tokenStore, uid)
		if err != nil {
			return err
		}

		if optUsr.IsNone() {
			continue
		}

		usr := optUsr.Unwrap()

		// migrate roles
		for _, id := range usr.Roles {
			err := m.rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: token.Namespace,
					Instance:  rebac.Instance(uid),
				},
				Relation: rebac.Member,
				Target: rebac.Entity{
					Namespace: role.Namespace,
					Instance:  rebac.Instance(id),
				},
			})
			if err != nil {
				return err
			}
		}

		// migrate groups
		for _, id := range usr.Groups {
			err := m.rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: token.Namespace,
					Instance:  rebac.Instance(uid),
				},
				Relation: rebac.Member,
				Target: rebac.Entity{
					Namespace: group.Namespace,
					Instance:  rebac.Instance(id),
				},
			})
			if err != nil {
				return err
			}
		}

		// migrate direct and global permissions
		for _, id := range usr.Permissions {
			err := m.rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: token.Namespace,
					Instance:  rebac.Instance(uid),
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

		// finally convert legacy resource-based permission assignments
		for res, ids := range usr.Resources {
			for _, id := range ids {
				err := m.rdb.Put(rebac.Triple{
					Source: rebac.Entity{
						Namespace: token.Namespace,
						Instance:  rebac.Instance(uid),
					},
					Relation: rebac.Relation(id),
					Target: rebac.Entity{
						Namespace: rebac.Namespace(res.Name),
						Instance:  rebac.Instance(res.ID),
					},
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
