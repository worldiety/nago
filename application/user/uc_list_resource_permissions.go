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
	"slices"
	"sync"
)

func NewListResourcePermissions(mutex *sync.Mutex, repo Repository) ListResourcePermissions {
	return func(subject AuditableUser, uid ID) iter.Seq2[ResourceWithPermissions, error] {
		if err := subject.Audit(PermListResourcePermissions); err != nil {
			return xiter.WithError[ResourceWithPermissions](err)
		}

		return func(yield func(ResourceWithPermissions, error) bool) {
			mutex.Lock()
			defer mutex.Unlock()

			optUsr, err := repo.FindByID(uid)
			if err != nil {
				yield(ResourceWithPermissions{}, err)
				return
			}

			if optUsr.IsNone() {
				// empty set of permissions
				return
			}

			usr := optUsr.Unwrap()
			for resource, ids := range usr.Resources {
				if !yield(ResourceWithPermissions{
					Permissions: slices.Values(ids),
					Resource:    resource,
				}, nil) {
					return
				}
			}
		}
	}
}
