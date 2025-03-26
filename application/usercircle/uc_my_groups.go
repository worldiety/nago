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
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewMyGroups(repo Repository, users user.UseCases, findGroupByID group.FindByID) MyGroups {
	return func(subject auth.Subject, circleId ID) iter.Seq2[group.Group, error] {

		circle, err := myCircle(repo, subject, circleId)
		if err != nil {
			return xiter.WithError[group.Group](err)
		}

		return func(yield func(group.Group, error) bool) {
			for _, gid := range circle.Groups {
				optGroup, err := findGroupByID(user.SU(), gid)
				if err != nil {
					if !yield(group.Group{}, err) {
						return
					}
					continue
				}

				if optGroup.IsNone() {
					continue
				}

				if !yield(optGroup.Unwrap(), nil) {
					return
				}
			}
		}
	}
}
