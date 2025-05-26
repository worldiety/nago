// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/pkg/std/tick"
	"go.wdy.de/nago/pkg/xstrings"
	"sync"
	"time"
)

func NewDisplayName(repo Repository, refreshInterval time.Duration) DisplayName {
	var mutex sync.Mutex
	var lastRefreshedAt time.Time
	cache := map[ID]Compact{}

	return func(uid ID) Compact {
		mutex.Lock()
		defer mutex.Unlock()

		if lastRefreshedAt.Add(refreshInterval).Before(tick.Now(tick.Minute)) {
			clear(cache)
		}

		info, ok := cache[uid]
		if ok {
			return info
		}

		var zero Compact

		optUsr, err := repo.FindByID(uid)
		if err != nil {
			return zero
		}

		if optUsr.IsNone() {
			return zero
		}

		usr := optUsr.Unwrap()
		info = Compact{
			ID:          uid,
			Avatar:      usr.Contact.Avatar,
			Displayname: xstrings.Join2(" ", usr.Contact.Firstname, usr.Contact.Lastname),
			Mail:        usr.Email,
			Valid:       usr.Enabled(),
		}

		cache[uid] = info

		return info
	}
}
