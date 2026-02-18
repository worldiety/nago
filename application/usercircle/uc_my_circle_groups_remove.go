// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"sync"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewMyCircleGroupsRemove(mutex *sync.Mutex, repo Repository, users user.UseCases, rdb *rebac.DB, usrRoles user.ListRoles, usrGroups user.ListGroups) MyCircleGroupsRemove {
	return func(subject auth.Subject, circleId ID, usrId user.ID, groups ...group.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, _, err := myCircleAndUser(repo, users.FindByID, usrRoles, usrGroups, subject, circleId, usrId)
		if err != nil {
			return err
		}

		for _, gid := range groups {
			err := rdb.Delete(rebac.Triple{
				Source: rebac.Entity{
					Namespace: group.Namespace,
					Instance:  rebac.Instance(gid),
				},
				Relation: rebac.Member,
				Target: rebac.Entity{
					Namespace: user.Namespace,
					Instance:  rebac.Instance(usrId),
				},
			})

			if err != nil {
				return err
			}
		}

		return nil
	}
}
