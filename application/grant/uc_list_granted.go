// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package grant

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewListGranted(repo Repository) ListGranted {
	return func(subject auth.Subject, res user.Resource) iter.Seq2[user.ID, error] {
		myID := NewID(res, subject.ID())

		// are we globally allowed?
		globalAllowed := subject.HasResourcePermission(repo.Name(), string(myID), PermListGranted)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasResourcePermission(res.Name, res.ID, PermListGranted)

		if !globalAllowed && !resAllowed {
			return xiter.WithError[user.ID](user.PermissionDeniedErr)
		}

		return func(yield func(user.ID, error) bool) {
			for granting, err := range repo.FindAllByPrefix(ID(res.Name)) {
				if err != nil {
					if !yield("", err) {
						return
					}

					continue
				}

				_, usr := granting.ID.Split()
				if !yield(usr, nil) {
					return
				}
			}
		}
	}
}
