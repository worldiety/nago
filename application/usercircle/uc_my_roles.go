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
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewMyRoles(repo Repository, users user.UseCases, findRoleByID role.FindByID) MyRoles {
	return func(subject auth.Subject, circleId ID) iter.Seq2[role.Role, error] {

		circle, err := myCircle(repo, subject, circleId)
		if err != nil {
			return xiter.WithError[role.Role](err)
		}

		return func(yield func(role.Role, error) bool) {
			for _, rid := range circle.Roles {
				optRole, err := findRoleByID(user.SU(), rid)
				if err != nil {
					if !yield(role.Role{}, err) {
						return
					}
					continue
				}

				if optRole.IsNone() {
					continue
				}

				if !yield(optRole.Unwrap(), nil) {
					return
				}
			}
		}
	}
}
