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

		v := newSessionImpl(id, repository, refresh)
		cache.Put(id, v)
		return v
	}
}
