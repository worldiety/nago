// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
)

func NewViewOf(users Repository, roles data.ReadRepository[role.Role, role.ID]) SubjectFromUser {
	return func(subject permission.Auditable, id ID) (std.Option[Subject], error) {
		// TODO not sure what permissions we need, this is only system anyway
		optUsr, err := users.FindByID(id)
		if err != nil {
			return std.None[Subject](), fmt.Errorf("failed to find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return std.None[Subject](), nil
		}

		return std.Some[Subject](newViewImpl(users, roles, optUsr.Unwrap())), nil
	}
}
