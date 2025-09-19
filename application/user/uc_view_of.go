// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewViewOf(users data.NotifyRepository[User, ID], roles data.ReadRepository[role.Role, role.ID]) SubjectFromUser {
	var cachedViews concurrent.RWMap[ID, *viewImpl]

	users.AddDeletedObserver(func(repository data.Repository[User, ID], deleted data.Deleted[ID]) error {
		cachedViews.Delete(deleted.ID)
		return nil
	})

	users.AddSavedObserver(func(repository data.Repository[User, ID], saved data.Saved[User, ID]) error {
		if v, ok := cachedViews.Get(saved.ID); ok {
			v.invalidate()
		}

		return nil
	})

	return func(subject permission.Auditable, id ID) (option.Opt[Subject], error) {
		// TODO not sure what permissions we need, this is only system anyway

		if usr, ok := cachedViews.Get(id); ok {
			return option.Some[Subject](usr), nil
		}

		optUsr, err := users.FindByID(id)
		if err != nil {
			return option.None[Subject](), fmt.Errorf("failed to find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return option.None[Subject](), nil
		}

		usr := newViewImpl(users, roles, optUsr.Unwrap())
		cachedViews.Put(id, usr)

		return option.Some[Subject](usr), nil
	}
}
