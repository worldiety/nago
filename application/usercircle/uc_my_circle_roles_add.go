// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"slices"
	"sync"
)

func NewMyCircleRolesAdd(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleRolesAdd {
	return func(subject auth.Subject, circleId ID, usrId user.ID, roles ...role.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, subject, circleId, usrId)
		if err != nil {
			return err
		}

		for _, rid := range roles {
			if !slices.Contains(usr.Roles, rid) {
				usr.Roles = append(usr.Roles, rid)
			}
		}

		return users.UpdateOtherRoles(user.SU(), usrId, usr.Roles)
	}
}
