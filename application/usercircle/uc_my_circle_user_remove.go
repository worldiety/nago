// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"sync"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewMyCircleUserRemove(mutex *sync.Mutex, repo Repository, users user.UseCases, usrRoles user.ListRoles, usrGroups user.ListGroups) MyCircleUserRemove {
	return func(subject auth.Subject, circleId ID, usrId user.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, usrRoles, usrGroups, subject, circleId, usrId)
		if err != nil {
			return err
		}

		return users.Delete(user.SU(), usr.ID)
	}
}
