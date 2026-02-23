// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewIsMyCircleMember(repo Repository, findUserByID user.FindByID, roles user.ListRoles, groups user.ListGroups) IsMyCircleMember {
	return func(subject auth.Subject, cid ID, other user.ID) (bool, error) {
		circle, err := myCircle(repo, subject, cid)
		if err != nil {
			return false, err
		}

		optUsr, err := findUserByID(user.SU(), other)
		if err != nil {
			return false, err
		}

		if optUsr.IsNone() {
			return false, nil
		}

		usr := optUsr.Unwrap()
		return circle.isMember(roles, groups, usr), nil

	}
}
