// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import "go.wdy.de/nago/pkg/std/concurrent"

func NewFindUserSessionByID(repository Repository, refresh RefreshNLS) FindUserSessionByID {
	var cache concurrent.RWMap[ID, *sessionImpl]

	return func(id ID) UserSession {
		if v, ok := cache.Get(id); ok {
			return v
		}

		// concurrent first accesses must share one instance, a scope keeps it for its lifetime
		v, _ := cache.LoadOrStore(id, newSessionImpl(id, repository, refresh))
		return v
	}
}
