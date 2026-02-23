// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewMyCircleRolesAdd(mutex *sync.Mutex, repo Repository, users user.UseCases, rdb *rebac.DB, usrRoles user.ListRoles, usrGroups user.ListGroups) MyCircleRolesAdd {
	return func(subject auth.Subject, circleId ID, usrId user.ID, roles ...role.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, _, err := myCircleAndUser(repo, users.FindByID, usrRoles, usrGroups, subject, circleId, usrId)
		if err != nil {
			return err
		}

		for _, rid := range roles {
			err := rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: role.Namespace,
					Instance:  rebac.Instance(rid),
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
