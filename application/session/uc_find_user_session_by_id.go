// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"log/slog"

	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewFindUserSessionByID(repository Repository, refresh RefreshNLS) FindUserSessionByID {
	find, _ := newFindUserSessionByID(repository, refresh)
	return find
}

// newFindUserSessionByID also returns the cache for tests. A cached instance only keeps the time of its last single
// sign-on refresh, see sessionImpl. Only sessions which exist are cached, because the id comes unchecked from a
// cookie: otherwise anybody could grow the cache without bounds by sending arbitrary ids.
func newFindUserSessionByID(repository Repository, refresh RefreshNLS) (FindUserSessionByID, *concurrent.RWMap[ID, *sessionImpl]) {
	var cache concurrent.RWMap[ID, *sessionImpl]

	return func(id ID) UserSession {
		if v, ok := cache.Get(id); ok {
			return v
		}

		v := newSessionImpl(id, repository, refresh)
		optSession, err := repository.FindByID(id)
		if err != nil {
			slog.Error("failed to find session by id", "err", err, "id", id)
			return v
		}

		if optSession.IsNone() {
			// an unknown id gets an instance nobody remembers, it still reads the repository on every access
			return v
		}

		// concurrent first accesses must share one instance, a scope keeps it for its lifetime
		v, _ = cache.LoadOrStore(id, v)
		return v
	}, &cache
}
