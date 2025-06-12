// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewListGrantedUsers(repo GrantingIndexRepository) ListGrantedUsers {
	return func(subject AuditableUser, res Resource) iter.Seq2[ID, error] {
		myID := NewGrantingKey(res, subject.ID())

		// are we globally allowed?
		globalAllowed := subject.HasResourcePermission(repo.Name(), string(myID), PermListGrantedUsers)

		// are we allowed for the specified resource+user?
		resAllowed := subject.HasResourcePermission(res.Name, res.ID, PermListGrantedUsers)

		if !globalAllowed && !resAllowed {
			return xiter.WithError[ID](PermissionDeniedErr)
		}

		return func(yield func(ID, error) bool) {
			for granting, err := range repo.FindAllByPrefix(GrantingKey(res.Name + "/" + res.ID)) {
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
