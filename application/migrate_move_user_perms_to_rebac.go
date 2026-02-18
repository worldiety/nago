// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/migration"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
)

type migrateUserPermsToReBAC struct {
	userStore blob.Store
	rdb       *rebac.DB
}

func newMigrateUserPermsToReBAC(store blob.Store, rdb *rebac.DB) migrateUserPermsToReBAC {
	return migrateUserPermsToReBAC{userStore: store, rdb: rdb}
}

func (m migrateUserPermsToReBAC) Version() migration.Version {
	return migration.NewVersion(2026, time.January, 14, 16, 31, "MoveUserPermsToReBAC")
}

func (m migrateUserPermsToReBAC) Migrate(ctx context.Context) error {
	// legacyUser is the last data model that still contains roles, groups, permissions and resources
	type legacyUser struct {
		ID          user.ID                                `json:"id"`
		Roles       []role.ID                              `json:"roles,omitempty"`       // roles may also contain inherited permissions
		Groups      []group.ID                             `json:"groups,omitempty"`      // groups may also contain inherited permissions
		Permissions []permission.ID                        `json:"permissions,omitempty"` // individual custom permissions
		Resources   map[legacyUserResource][]permission.ID `json:"resources,omitempty"`
	}

	for uid, err := range m.userStore.List(ctx, blob.ListOptions{}) {
		if err != nil {
			return err
		}
		optUsr, err := json.Get[legacyUser](m.userStore, uid)
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
					Namespace: user.Namespace,
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
					Namespace: user.Namespace,
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
					Namespace: user.Namespace,
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
						Namespace: user.Namespace,
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

type legacyUserResource struct {
	// ID is the string version of the root aggregate or entity key used in the named Store or Repository.
	// If ID is empty, all values from the Named Store or Repository are applicable.
	ID string

	// Name of the Store or Repository
	Name string
}

func (r *legacyUserResource) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		r.Name = ""
		r.ID = ""
	}

	str, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	tokens := strings.SplitN(str, "/", 2)
	if len(tokens) != 2 {
		return fmt.Errorf("invalid json format for resource: %s", str)
	}

	r.Name = tokens[0]
	r.ID = tokens[1]
	return nil
}
