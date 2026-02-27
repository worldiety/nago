// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
)

// RDB returns the nago ReBAC (relation-based access control) database. Even though there is a separate module,
// the rebac system is always available, and the module is only required if you want the admin user interface for it.
// The default resolvers are
//   - users which are members of a role resolve to the assigned role relations
func (c *Configurator) RDB() (*rebac.DB, error) {
	if c.rdb == nil {
		store, err := c.EntityStore("nago.rebac")
		if err != nil {
			return nil, err
		}

		db, err := rebac.NewDB(store)
		if err != nil {
			return nil, err
		}

		// automatically resolve role relations by user memberships
		db.AddResolver(rebac.NewSourceMemberResolver(user.Namespace, role.Namespace))

		c.rdb = db

	}

	return c.rdb, nil
}
