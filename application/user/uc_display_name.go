// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"sync"
	"time"

	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std/tick"
	"go.wdy.de/nago/pkg/xstrings"
)

// NewDisplayName caches the compact information of each requested user and drops the whole cache once the
// refreshInterval has passed. If the repository is a [data.NotifyRepository], a saved or deleted user is evicted at
// once as well, so that a changed name, avatar, mail or account status is visible with the next lookup and not only
// after the refreshInterval. The eviction happens synchronously within the repository call, thus before the use
// cases publish their domain events like [ContactUpdated] or [Deleted]: observers of these events already read the
// new values. A user whose account is only enabled until a point in time is not served from the cache beyond that
// point, so that [Compact.Valid] turns false in time.
func NewDisplayName(repo Repository, refreshInterval time.Duration) DisplayName {
	return newDisplayName(repo, refreshInterval, func() time.Time { return tick.Now(tick.Minute) })
}

// displayNameEntry is a cached [Compact] together with the point in time, at which it becomes outdated by itself.
// A zero expiresAt never expires by itself.
type displayNameEntry struct {
	info      Compact
	expiresAt time.Time
}

func newDisplayName(repo Repository, refreshInterval time.Duration, now func() time.Time) DisplayName {
	var mutex sync.Mutex
	var lastRefreshedAt time.Time
	cache := map[ID]displayNameEntry{}

	if notifyRepo, ok := repo.(data.NotifyRepository[User, ID]); ok {
		// note: we never detach, because the cache lives as long as the use cases do.
		notifyRepo.AddSavedObserver(func(_ data.Repository[User, ID], saved data.Saved[User, ID]) error {
			mutex.Lock()
			defer mutex.Unlock()

			delete(cache, saved.ID)
			return nil
		})

		notifyRepo.AddDeletedObserver(func(_ data.Repository[User, ID], deleted data.Deleted[ID]) error {
			mutex.Lock()
			defer mutex.Unlock()

			delete(cache, deleted.ID)
			return nil
		})
	}

	return func(uid ID) Compact {
		mutex.Lock()
		defer mutex.Unlock()

		if t := now(); lastRefreshedAt.Add(refreshInterval).Before(t) {
			clear(cache)
			lastRefreshedAt = t
		}

		entry, ok := cache[uid]
		// note: the exact clock is only asked for the rare accounts with a time limit
		if ok && (entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt)) {
			return entry.info
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
		entry = displayNameEntry{
			info: Compact{
				ID:          uid,
				Avatar:      usr.Contact.Avatar,
				Displayname: xstrings.Join2(" ", usr.Contact.Firstname, usr.Contact.Lastname),
				Mail:        usr.Email,
				Valid:       usr.Enabled(),
			},
		}

		if status, ok := usr.Status.(EnabledUntil); ok && entry.info.Valid {
			entry.expiresAt = status.ValidUntil
		}

		cache[uid] = entry

		return entry.info
	}
}
