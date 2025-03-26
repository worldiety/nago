// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"slices"
	"sync"
)

func NewMyCircleGroupsAdd(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleGroupsAdd {
	return func(subject auth.Subject, circleId ID, usrId user.ID, groups ...group.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, subject, circleId, usrId)
		if err != nil {
			return err
		}

		for _, rid := range groups {
			if !slices.Contains(usr.Groups, rid) {
				usr.Groups = append(usr.Groups, rid)
			}
		}

		return users.UpdateOtherGroups(user.SU(), usrId, usr.Groups)
	}
}
